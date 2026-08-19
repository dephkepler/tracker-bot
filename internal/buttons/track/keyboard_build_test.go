package track

import (
	"testing"
	"tracker-bot/internal/i18n"
)

func TestFormatParseTimerButton_RoundTrip(t *testing.T) {
	for _, lang := range i18n.All {
		for _, prefix := range []string{TrackTimerActivatePrefix, TrackTimerDeletePrefix} {
			for _, minutes := range []int{1, 15, 30, 360} {
				text := FormatTimerButton(lang, prefix, minutes)
				got, ok := ParseTimerButtonMinutes(lang, text, prefix)
				if !ok {
					t.Fatalf("[%s] ParseTimerButtonMinutes(%q, %q): ok = false, want true", lang, text, prefix)
				}
				if got != minutes {
					t.Fatalf("[%s] ParseTimerButtonMinutes(%q, %q) = %d, want %d", lang, text, prefix, got, minutes)
				}
			}
		}
	}
}

// TestParseTimerButtonMinutes_FallsBackToDefault checks a button rendered
// in English (e.g. from before the user picked a language, or a
// not-yet-converted screen) still parses correctly under a non-English
// ctx.Language — the same cross-phase safety net as i18n.Key.
func TestParseTimerButtonMinutes_FallsBackToDefault(t *testing.T) {
	text := FormatTimerButton(i18n.Default, TrackTimerActivatePrefix, 45)
	for _, lang := range i18n.All {
		got, ok := ParseTimerButtonMinutes(lang, text, TrackTimerActivatePrefix)
		if !ok {
			t.Fatalf("[%s] ParseTimerButtonMinutes(%q): ok = false, want true", lang, text)
		}
		if got != 45 {
			t.Fatalf("[%s] ParseTimerButtonMinutes(%q) = %d, want 45", lang, text, got)
		}
	}
}

func TestParseTimerButtonMinutes_RejectsNonMatching(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		prefix string
	}{
		{"wrong prefix", FormatTimerButton(i18n.EN, TrackTimerActivatePrefix, 15), TrackTimerDeletePrefix},
		{"no prefix at all", "15 min", TrackTimerActivatePrefix},
		{"missing suffix", "⏱ 15", TrackTimerActivatePrefix},
		{"not a number", "⏱ fifteen min", TrackTimerActivatePrefix},
		{"zero", "⏱ 0 min", TrackTimerActivatePrefix},
		{"negative", "⏱ -5 min", TrackTimerActivatePrefix},
		// The static "🗑 Delete Timer" button shares the "🗑 " prefix with
		// delete-picker buttons ("🗑 15 min") — it must never be mistaken
		// for one because it doesn't end in a minutes unit at all.
		{"delete-timer button text itself", i18n.T(i18n.EN, i18n.KeyTrackButtonTimerDelete), TrackTimerDeletePrefix},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := ParseTimerButtonMinutes(i18n.EN, tc.text, tc.prefix); ok {
				t.Fatalf("ParseTimerButtonMinutes(%q, %q): ok = true, want false", tc.text, tc.prefix)
			}
		})
	}
}
