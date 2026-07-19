// Package player streams internet radio by decoding it with ffmpeg and
// playing the resulting PCM through the system audio device via oto.
package player

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

const (
	sampleRate   = 44100
	channelCount = 2

	reconnectBaseDelay = 1 * time.Second
	reconnectMaxDelay  = 30 * time.Second
)

// State describes the current playback state.
type State int

const (
	// StateStopped means nothing is playing.
	StateStopped State = iota
	// StateConnecting means the stream is being opened.
	StateConnecting
	// StatePlaying means audio is playing.
	StatePlaying
	// StatePaused means playback is paused.
	StatePaused
	// StateError means the last attempt failed.
	StateError
)

// Status is a snapshot of the player state reported to observers.
type Status struct {
	State   State
	Station string
	Err     error
}

// StatusFunc receives player status updates. It may be called from a
// background goroutine, so implementations must be safe to call from any
// goroutine.
type StatusFunc func(Status)

// Player controls playback of a single station at a time.
type Player struct {
	ctx        *oto.Context
	ffmpegPath string
	onStatus   StatusFunc

	mu         sync.Mutex
	generation int
	station    string
	volume     float64
	current    *oto.Player
	cmd        *exec.Cmd
	stdout     io.ReadCloser
	state      State
}

// New creates a Player. It initializes the shared audio context, which can
// only be created once per process. ffmpegPath is the ffmpeg binary used to
// decode streams.
func New(ffmpegPath string, volume float64, onStatus StatusFunc) (*Player, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, err
	}
	<-ready

	return &Player{
		ctx:        ctx,
		ffmpegPath: ffmpegPath,
		onStatus:   onStatus,
		volume:     clampVolume(volume),
		state:      StateStopped,
	}, nil
}

func clampVolume(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// FindFFmpeg locates a usable ffmpeg binary, preferring one bundled next to
// the running executable and falling back to one found on PATH.
func FindFFmpeg() (string, error) {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, candidate := range []string{
			filepath.Join(dir, "ffmpeg"),
			filepath.Join(dir, "bin", "ffmpeg"),
		} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
				return candidate, nil
			}
		}
	}
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path, nil
	}
	return "", errors.New("ffmpeg не найден")
}

func (p *Player) notify(status Status) {
	if p.onStatus != nil {
		p.onStatus(status)
	}
}

// Play starts playing the given station. If the station is already loaded and
// paused, it resumes instead of reconnecting.
func (p *Player) Play(name, url string) {
	p.mu.Lock()
	if p.state == StatePaused && p.station == name && p.current != nil {
		p.current.Play()
		p.state = StatePlaying
		p.mu.Unlock()
		p.notify(Status{State: StatePlaying, Station: name})
		return
	}
	p.mu.Unlock()

	p.start(name, url)
}

func (p *Player) start(name, url string) {
	p.mu.Lock()
	p.teardownLocked()
	p.generation++
	generation := p.generation
	p.station = name
	p.state = StateConnecting
	volume := p.volume
	p.mu.Unlock()

	p.notify(Status{State: StateConnecting, Station: name})
	go p.connect(generation, name, url, volume, 0)
}

func (p *Player) connect(generation int, name, url string, volume float64, attempt int) {
	p.mu.Lock()
	volume = p.volume
	p.mu.Unlock()

	cmd := exec.Command(p.ffmpegPath,
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-reconnect", "1",
		"-reconnect_at_eof", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "30",
		"-i", url,
		"-vn", "-sn", "-dn",
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ar", "44100",
		"-ac", "2",
		"pipe:1",
	)
	cmd.Stderr = nil

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.fail(generation, name, url, volume, attempt, err)
		return
	}
	if err := cmd.Start(); err != nil {
		p.fail(generation, name, url, volume, attempt, err)
		return
	}

	otoPlayer := p.ctx.NewPlayer(stdout)
	otoPlayer.SetVolume(volume)
	otoPlayer.Play()

	p.mu.Lock()
	if generation != p.generation {
		p.mu.Unlock()
		otoPlayer.Close()
		_ = stdout.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return
	}
	p.current = otoPlayer
	p.cmd = cmd
	p.stdout = stdout
	p.state = StatePlaying
	p.mu.Unlock()

	p.notify(Status{State: StatePlaying, Station: name})

	_ = cmd.Wait()

	p.mu.Lock()
	if generation != p.generation {
		p.mu.Unlock()
		return
	}
	// The process ended on its own while we still expected playback.
	if p.current != nil {
		p.current.Pause()
		_ = p.current.Close()
		p.current = nil
	}
	if p.stdout != nil {
		_ = p.stdout.Close()
		p.stdout = nil
	}
	p.cmd = nil
	p.state = StateConnecting
	p.mu.Unlock()

	p.notify(Status{State: StateConnecting, Station: name})
	p.reconnect(generation, name, url, volume, attempt+1)
}

func (p *Player) reconnect(generation int, name, url string, volume float64, attempt int) {
	delay := reconnectBaseDelay << uint(min(attempt, 5))
	if delay > reconnectMaxDelay {
		delay = reconnectMaxDelay
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	<-timer.C

	p.mu.Lock()
	stale := generation != p.generation
	p.mu.Unlock()
	if stale {
		return
	}
	p.connect(generation, name, url, volume, attempt)
}

func (p *Player) fail(generation int, name, url string, volume float64, attempt int, cause error) {
	p.mu.Lock()
	if generation != p.generation {
		p.mu.Unlock()
		return
	}
	p.state = StateError
	p.mu.Unlock()
	p.notify(Status{State: StateError, Station: name, Err: cause})
	p.reconnect(generation, name, url, volume, attempt+1)
}

// Pause pauses playback while keeping the stream loaded.
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current == nil || p.state != StatePlaying {
		return
	}
	p.current.Pause()
	p.state = StatePaused
	name := p.station
	go p.notify(Status{State: StatePaused, Station: name})
}

// Stop tears down playback completely.
func (p *Player) Stop() {
	p.mu.Lock()
	p.teardownLocked()
	p.state = StateStopped
	p.station = ""
	p.mu.Unlock()
	p.notify(Status{State: StateStopped})
}

func (p *Player) teardownLocked() {
	p.generation++
	if p.current != nil {
		p.current.Pause()
		_ = p.current.Close()
		p.current = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}
	if p.stdout != nil {
		_ = p.stdout.Close()
		p.stdout = nil
	}
}

// SetVolume updates the playback volume (0..1).
func (p *Player) SetVolume(volume float64) {
	volume = clampVolume(volume)
	p.mu.Lock()
	p.volume = volume
	if p.current != nil {
		p.current.SetVolume(volume)
	}
	p.mu.Unlock()
}

// Volume returns the current volume (0..1).
func (p *Player) Volume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

// State returns the current playback state.
func (p *Player) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Close stops playback and releases resources.
func (p *Player) Close() {
	p.mu.Lock()
	p.teardownLocked()
	p.state = StateStopped
	p.mu.Unlock()
}
