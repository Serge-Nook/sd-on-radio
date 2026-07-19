package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidHTTPURL(t *testing.T) {
	cases := map[string]bool{
		"http://example.com/stream":      true,
		"HTTP://example.com/stream":      true,
		"https://example.com:8000/s.mp3": true,
		"ftp://example.com/stream":       false,
		"example.com/stream":             false,
		"":                               false,
		"http://":                        false,
	}
	for input, want := range cases {
		if got := IsValidHTTPURL(input); got != want {
			t.Errorf("IsValidHTTPURL(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestReadWrongFieldTypesFallsBackIndividually(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	data := `{
		"stations": [
			{"name": "Good", "url": "https://example.com/s"},
			{"name": 42, "url": "https://example.com/bad"}
		],
		"last_station": 42,
		"volume": "loud"
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Stations) != 1 || got.Stations[0].Name != "Good" {
		t.Errorf("valid station was not preserved: %+v", got.Stations)
	}
	if got.LastStation != "" || got.Volume != defaultVolume {
		t.Errorf("invalid fields did not use defaults: %+v", got)
	}
}

func TestNormalizeStation(t *testing.T) {
	station, ok := NormalizeStation(Station{Name: "  My Radio  ", URL: "  http://example.com/s  "})
	if !ok {
		t.Fatal("expected station to be valid")
	}
	if station.Name != "My Radio" || station.URL != "http://example.com/s" {
		t.Errorf("unexpected trim result: %+v", station)
	}

	if _, ok := NormalizeStation(Station{Name: "", URL: "http://example.com"}); ok {
		t.Error("empty name should be invalid")
	}
	if _, ok := NormalizeStation(Station{Name: "ok", URL: "notaurl"}); ok {
		t.Error("invalid url should be invalid")
	}
}

func TestNormalizeClampsVolume(t *testing.T) {
	if v := Normalize(Settings{Volume: 5}).Volume; v != 1 {
		t.Errorf("volume not clamped high: %v", v)
	}
	if v := Normalize(Settings{Volume: -2}).Volume; v != 0 {
		t.Errorf("volume not clamped low: %v", v)
	}
}

func TestNormalizeDropsInvalidStations(t *testing.T) {
	in := Settings{Stations: []Station{
		{Name: "Good", URL: "http://example.com/s"},
		{Name: "Bad", URL: "ftp://example.com"},
	}}
	got := Normalize(in)
	if len(got.Stations) != 1 || got.Stations[0].Name != "Good" {
		t.Errorf("invalid stations not dropped: %+v", got.Stations)
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.json")

	want := Settings{
		Stations:    []Station{{Name: "Radio", URL: "https://example.com/s"}},
		LastStation: "Radio",
		Volume:      0.5,
	}
	if _, err := Write(path, want); err != nil {
		t.Fatalf("write: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v, want 0600", info.Mode().Perm())
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.LastStation != "Radio" || got.Volume != 0.5 || len(got.Stations) != 1 {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestReadMissingCreatesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Volume != defaultVolume {
		t.Errorf("default volume = %v, want %v", got.Volume, defaultVolume)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("default file not created: %v", err)
	}
}

func TestReadCorruptRecoversDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Volume != defaultVolume {
		t.Errorf("corrupt file not recovered: %+v", got)
	}
}

func TestReadMissingVolumeUsesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"stations":[],"last_station":""}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Volume != defaultVolume {
		t.Errorf("missing volume = %v, want %v", got.Volume, defaultVolume)
	}
}
