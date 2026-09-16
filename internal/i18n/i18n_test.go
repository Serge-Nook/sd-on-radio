package i18n

import "testing"

func TestAllKeysTranslatedInEveryLanguage(t *testing.T) {
	for key, entry := range messages {
		for _, lang := range Languages() {
			if entry[lang] == "" {
				t.Errorf("key %q has no %s translation", key, lang)
			}
		}
	}
}

func TestTFallsBack(t *testing.T) {
	if got := T(Language("xx"), "settings"); got != messages["settings"][Russian] {
		t.Errorf("unknown language did not fall back to Russian: %q", got)
	}
	if got := T(English, "missing_key"); got != "missing_key" {
		t.Errorf("unknown key did not return key: %q", got)
	}
}

func TestNativeNameRoundTrip(t *testing.T) {
	for _, lang := range Languages() {
		got, ok := FromNativeName(NativeName(lang))
		if !ok || got != lang {
			t.Errorf("round trip failed for %s: %s %v", lang, got, ok)
		}
	}
	if !IsValid(Chinese) || IsValid(Language("xx")) {
		t.Error("IsValid mismatch")
	}
}
