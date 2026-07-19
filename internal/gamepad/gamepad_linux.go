//go:build linux

package gamepad

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const (
	eventButton = 0x01
	eventAxis   = 0x02
	eventInit   = 0x80

	axisThreshold = 20000
)

// event mirrors the Linux joystick js_event structure (8 bytes).
type event struct {
	Time   uint32
	Value  int16
	Type   uint8
	Number uint8
}

// Reader polls a Linux joystick device and dispatches navigation actions.
type Reader struct {
	file *os.File
	done chan struct{}
}

// Start opens the first available joystick and begins reading in the
// background. It returns nil (not an error) when no controller is available so
// the application keeps working without one.
func Start(handler Handler) *Reader {
	path := firstDevice()
	if path == "" {
		return nil
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil
	}
	_ = unix.SetNonblock(int(file.Fd()), false)

	r := &Reader{file: file, done: make(chan struct{})}
	go r.loop(handler)
	return r
}

func firstDevice() string {
	matches, err := filepath.Glob("/dev/input/js*")
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func (r *Reader) loop(handler Handler) {
	buf := make([]byte, 8)
	var lastVertical int16
	for {
		select {
		case <-r.done:
			return
		default:
		}

		if _, err := io.ReadFull(r.file, buf); err != nil {
			return
		}

		var ev event
		ev.Time = binary.LittleEndian.Uint32(buf[0:4])
		ev.Value = int16(binary.LittleEndian.Uint16(buf[4:6]))
		ev.Type = buf[6]
		ev.Number = buf[7]

		if ev.Type&eventInit != 0 {
			continue
		}

		switch ev.Type {
		case eventButton:
			if ev.Value != 1 {
				continue
			}
			switch ev.Number {
			case 0: // South / A
				handler(ActionSelect)
			case 1: // East / B
				handler(ActionBack)
			case 2, 3: // West / North
				handler(ActionPlayPause)
			case 4: // Left shoulder
				handler(ActionVolumeDown)
			case 5: // Right shoulder
				handler(ActionVolumeUp)
			}
		case eventAxis:
			// Treat the vertical axis of the D-pad / left stick as list navigation.
			if ev.Number != 1 && ev.Number != 7 {
				continue
			}
			switch {
			case ev.Value < -axisThreshold && lastVertical >= -axisThreshold:
				handler(ActionUp)
			case ev.Value > axisThreshold && lastVertical <= axisThreshold:
				handler(ActionDown)
			}
			lastVertical = ev.Value
		}
	}
}

// Stop stops reading and closes the device.
func (r *Reader) Stop() {
	if r == nil {
		return
	}
	close(r.done)
	_ = r.file.Close()
}
