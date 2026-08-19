package track

import (
	"strings"
	"testing"
	"time"
	"tracker-bot/internal/i18n"
)

// TestTrackHeatmapText_TallyCountsOnlyPastDays checks the caption's tally
// only counts days up to and including "today" — upcoming days in the
// still-incomplete current week must not inflate the denominator.
func TestTrackHeatmapText_TallyCountsOnlyPastDays(t *testing.T) {
	monday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	today := monday // today is Monday itself — rest of week is upcoming

	tracked := map[string]bool{monday.Format("2006-01-02"): true}

	got := TrackHeatmapText(i18n.EN, monday, today, tracked, 1)
	if !strings.Contains(got, "1 of 1 days tracked") {
		t.Fatalf("output = %q, want tally \"1 of 1 days tracked\" (only Monday counts as past)", got)
	}
}

// TestTrackHeatmapInlineMenu_MarksTrackedMissedAndUpcoming checks the three
// cell states show up as the right emoji buttons: 🟩 tracked, ⬛ missed
// (past, not tracked), ⬜ upcoming (after "today", inert).
func TestTrackHeatmapInlineMenu_MarksTrackedMissedAndUpcoming(t *testing.T) {
	monday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	today := monday.AddDate(0, 0, 2) // Wednesday — Thu-Sun of this week are "upcoming"

	tracked := map[string]bool{
		monday.Format("2006-01-02"): true, // Monday tracked
		// Tuesday missed (not in map), Wednesday (today) also missed
	}

	menu := TrackHeatmapInlineMenu(i18n.EN, monday, today, tracked, 1)
	// Row 0 is the weekday header, row 1 is the only week.
	if len(menu.InlineKeyboard) < 2 {
		t.Fatalf("menu has %d rows, want at least a header + one week row", len(menu.InlineKeyboard))
	}
	week := menu.InlineKeyboard[1]
	if len(week) != 7 {
		t.Fatalf("week row has %d buttons, want 7", len(week))
	}

	want := []string{"🟩", "⬛", "⬛", "⬜", "⬜", "⬜", "⬜"}
	for i, btn := range week {
		if btn.Text != want[i] {
			t.Errorf("cell[%d] = %q, want %q", i, btn.Text, want[i])
		}
	}

	// Tracked/missed days must carry a real callback; upcoming days must be inert.
	if week[0].CallbackData == nil || *week[0].CallbackData != TrackCBHeatmapDay+monday.Format("2006-01-02") {
		t.Errorf("Monday (tracked) callback = %v, want %s", week[0].CallbackData, TrackCBHeatmapDay+monday.Format("2006-01-02"))
	}
	if week[3].CallbackData == nil || *week[3].CallbackData != "noop" {
		t.Errorf("Thursday (upcoming) callback = %v, want noop", week[3].CallbackData)
	}
}
