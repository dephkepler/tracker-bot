package track

import "testing"

func TestFormatParseTimerButton_RoundTrip(t *testing.T) {
	for _, prefix := range []string{TrackTimerActivatePrefix, TrackTimerDeletePrefix} {
		for _, minutes := range []int{1, 15, 30, 360} {
			text := FormatTimerButton(prefix, minutes)
			got, ok := ParseTimerButtonMinutes(text, prefix)
			if !ok {
				t.Fatalf("ParseTimerButtonMinutes(%q, %q): ok = false, want true", text, prefix)
			}
			if got != minutes {
				t.Fatalf("ParseTimerButtonMinutes(%q, %q) = %d, want %d", text, prefix, got, minutes)
			}
		}
	}
}

func TestParseTimerButtonMinutes_RejectsNonMatching(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		prefix string
	}{
		{"wrong prefix", FormatTimerButton(TrackTimerActivatePrefix, 15), TrackTimerDeletePrefix},
		{"no prefix at all", "15 min", TrackTimerActivatePrefix},
		{"missing suffix", "⏱ 15", TrackTimerActivatePrefix},
		{"not a number", "⏱ fifteen min", TrackTimerActivatePrefix},
		{"zero", "⏱ 0 min", TrackTimerActivatePrefix},
		{"negative", "⏱ -5 min", TrackTimerActivatePrefix},
		// The static "🗑 Delete Timer" button shares the "🗑 " prefix with
		// delete-picker buttons ("🗑 15 min") — it must never be
		// mistaken for one because it doesn't end in " min".
		{"delete-timer button text itself", TrackButtonTimerDelete, TrackTimerDeletePrefix},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := ParseTimerButtonMinutes(tc.text, tc.prefix); ok {
				t.Fatalf("ParseTimerButtonMinutes(%q, %q): ok = true, want false", tc.text, tc.prefix)
			}
		})
	}
}
