package stations

import "testing"

func TestBuiltInExcludesLoveRadio(t *testing.T) {
	for _, station := range BuiltIn() {
		if station.Name == "Love Radio" {
			t.Fatalf("Love Radio must not be a built-in station")
		}
	}
}

func TestBuiltInStationsAreValid(t *testing.T) {
	list := BuiltIn()
	if len(list) != 7 {
		t.Fatalf("expected 7 built-in stations, got %d", len(list))
	}
	seen := map[string]bool{}
	for _, station := range list {
		if station.Name == "" || station.URL == "" {
			t.Errorf("station with empty field: %+v", station)
		}
		if seen[station.Name] {
			t.Errorf("duplicate station name: %s", station.Name)
		}
		seen[station.Name] = true
	}
}
