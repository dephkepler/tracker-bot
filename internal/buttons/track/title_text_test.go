package track

import (
	"strings"
	"testing"
	"time"
	"tracker-bot/internal/i18n"
)

// TestTrackHeatmapText_MarksTrackedMissedAndUpcoming checks the three cell
// states: 🟩 a day with tracked time, ⬛ a past day with none, ⬜ a day after
// "today" (the still-incomplete current week).
func TestTrackHeatmapText_MarksTrackedMissedAndUpcoming(t *testing.T) {
	monday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC) // a real Monday
	today := monday.AddDate(0, 0, 2)                       // Wednesday — Thu-Sun of this week are "upcoming"

	tracked := map[string]bool{
		monday.Format("2006-01-02"): true, // Monday tracked
		// Tuesday missed (not in map)
	}

	got := TrackHeatmapText(i18n.EN, monday, today, tracked, 1)

	lines := strings.Split(got, "\n")
	// Header title, blank, weekday header, one week row, blank, legend, tally.
	var weekRow string
	for _, l := range lines {
		if strings.Contains(l, "🟩") || strings.Contains(l, "⬛") {
			weekRow = l
			break
		}
	}
	if weekRow == "" {
		t.Fatalf("output = %q, want a week row containing 🟩/⬛", got)
	}

	cells := strings.Fields(weekRow)
	if len(cells) != 7 {
		t.Fatalf("week row has %d cells, want 7: %q", len(cells), weekRow)
	}
	want := []string{"🟩", "⬛", "🟩", "⬜", "⬜", "⬜", "⬜"}
	// Wednesday (today) counts as tracked-if-present too, but we didn't mark
	// it — it's "past" (today itself), so expect ⬛ there; index 2 = Wed.
	want[2] = "⬛"
	for i := range want {
		if cells[i] != want[i] {
			t.Errorf("cell[%d] = %q, want %q (row: %q)", i, cells[i], want[i], weekRow)
		}
	}
}

func TestTrackHeatmapText_TallyCountsOnlyPastDays(t *testing.T) {
	monday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	today := monday // today is Monday itself — rest of week is upcoming

	tracked := map[string]bool{monday.Format("2006-01-02"): true}

	got := TrackHeatmapText(i18n.EN, monday, today, tracked, 1)
	if !strings.Contains(got, "1 of 1 days tracked") {
		t.Fatalf("output = %q, want tally \"1 of 1 days tracked\" (only Monday counts as past)", got)
	}
}
