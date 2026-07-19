// Package settings persists user preferences and custom stations as JSON.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxStationNameLen = 80
	maxStationURLLen  = 2048
	defaultVolume     = 0.8
)

// Station is a single radio station entry.
type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Settings holds the persisted application state.
type Settings struct {
	Stations    []Station `json:"stations"`
	LastStation string    `json:"last_station"`
	Volume      float64   `json:"volume"`
}

// Default returns the built-in default settings.
func Default() Settings {
	return Settings{
		Stations:    []Station{},
		LastStation: "",
		Volume:      defaultVolume,
	}
}

// IsValidHTTPURL reports whether value is a well-formed http(s) URL with a host.
func IsValidHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}
	return parsed.Host != ""
}

// NormalizeStation trims and validates a station, returning ok=false when invalid.
func NormalizeStation(station Station) (Station, bool) {
	name := strings.TrimSpace(station.Name)
	streamURL := strings.TrimSpace(station.URL)

	if name == "" || utf8.RuneCountInString(name) > maxStationNameLen {
		return Station{}, false
	}
	if len(streamURL) > maxStationURLLen || !IsValidHTTPURL(streamURL) {
		return Station{}, false
	}
	return Station{Name: name, URL: streamURL}, true
}

func clampVolume(volume float64) float64 {
	if math.IsNaN(volume) || math.IsInf(volume, 0) {
		return defaultVolume
	}
	if volume < 0 {
		return 0
	}
	if volume > 1 {
		return 1
	}
	return volume
}

func truncateName(name string) string {
	runes := []rune(name)
	if len(runes) > maxStationNameLen {
		return string(runes[:maxStationNameLen])
	}
	return name
}

// Normalize sanitizes arbitrary input into valid Settings.
func Normalize(value Settings) Settings {
	stations := make([]Station, 0, len(value.Stations))
	for _, station := range value.Stations {
		if normalized, ok := NormalizeStation(station); ok {
			stations = append(stations, normalized)
		}
	}
	return Settings{
		Stations:    stations,
		LastStation: truncateName(value.LastStation),
		Volume:      clampVolume(value.Volume),
	}
}

// Path returns the default settings file location.
func Path() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".local", "share", "sd-on-radio", "settings.json"), nil
}

// Read loads settings from path, recreating defaults for missing or corrupt files.
func Read(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Write(path, Default())
		}
		return Settings{}, err
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return Write(path, Default())
	}
	source, ok := value.(map[string]any)
	if !ok {
		return Default(), nil
	}

	result := Default()
	if rawStations, ok := source["stations"].([]any); ok {
		result.Stations = make([]Station, 0, len(rawStations))
		for _, rawStation := range rawStations {
			object, ok := rawStation.(map[string]any)
			if !ok {
				continue
			}
			name, nameOK := object["name"].(string)
			streamURL, urlOK := object["url"].(string)
			if !nameOK || !urlOK {
				continue
			}
			result.Stations = append(result.Stations, Station{Name: name, URL: streamURL})
		}
	}
	if lastStation, ok := source["last_station"].(string); ok {
		result.LastStation = lastStation
	}
	if volume, ok := source["volume"].(float64); ok {
		result.Volume = volume
	}
	return Normalize(result), nil
}

// Write atomically persists normalized settings and returns what was stored.
func Write(path string, value Settings) (Settings, error) {
	normalized := Normalize(value)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Settings{}, err
	}

	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return Settings{}, err
	}
	data = append(data, '\n')

	tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return Settings{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return Settings{}, err
	}
	return normalized, nil
}
