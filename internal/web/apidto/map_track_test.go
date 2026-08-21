package apidto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"tracker-bot/internal/models"
)

func TestSecondsRoundsDown(t *testing.T) {
	cases := map[time.Duration]int64{
		0:                                  0,
		90 * time.Second:                   90,
		2 * time.Hour:                      7200,
		time.Minute + 500*time.Millisecond: 60,
	}
	for d, want := range cases {
		if got := Seconds(d); got != want {
			t.Fatalf("Seconds(%v) = %d, want %d", d, got, want)
		}
	}
}

// A duration must never reach the wire as a raw time.Duration: encoding/json
// writes it as int64 nanoseconds, which is a chart six orders of magnitude too
// tall with nothing erroring.
func TestDurationsCrossTheWireAsSeconds(t *testing.T) {
	out := FromActivityStats([]models.ActivityDurationStat{
		{ActivityID: 1, Name: "Go", Duration: 90 * time.Minute, Sessions: 3},
	}, 90*time.Minute)

	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"seconds":5400`) {
		t.Fatalf("expected seconds, got %s", raw)
	}
	if strings.Contains(string(raw), "5400000000000") {
		t.Fatalf("nanoseconds leaked to the wire: %s", raw)
	}
}

func TestSharePercent(t *testing.T) {
	total := 100 * time.Minute
	stats := []models.ActivityDurationStat{
		{ActivityID: 1, Duration: 50 * time.Minute},
		{ActivityID: 2, Duration: 33 * time.Minute},
		{ActivityID: 3, Duration: 1 * time.Minute},
	}

	out := FromActivityStats(stats, total)
	want := []float64{50, 33, 1}
	for i, w := range want {
		if out[i].SharePercent != w {
			t.Fatalf("share[%d] = %v, want %v", i, out[i].SharePercent, w)
		}
	}
}

func TestSharePercentRoundsToOneDecimal(t *testing.T) {
	// A third of the period: 33.333… must not arrive as a float with a
	// twelve-digit tail, which is what an un-rounded division produces.
	out := FromActivityStats([]models.ActivityDurationStat{
		{ActivityID: 1, Duration: time.Hour},
	}, 3*time.Hour)

	if out[0].SharePercent != 33.3 {
		t.Fatalf("share = %v, want 33.3", out[0].SharePercent)
	}
}

// An empty period must not divide by zero.
func TestSharePercentOfAnEmptyPeriod(t *testing.T) {
	out := FromActivityStats([]models.ActivityDurationStat{
		{ActivityID: 1, Duration: 0},
	}, 0)

	if out[0].SharePercent != 0 {
		t.Fatalf("share = %v, want 0", out[0].SharePercent)
	}
}

// Shares are computed against the total the caller supplies, not against the
// sum of the rows: a top-N breakdown is a slice of the period, and re-deriving
// the total from it would inflate every share to fill 100%.
func TestSharePercentUsesTheGivenTotalNotTheRowSum(t *testing.T) {
	out := FromActivityStats([]models.ActivityDurationStat{
		{ActivityID: 1, Duration: 30 * time.Minute},
		{ActivityID: 2, Duration: 20 * time.Minute},
	}, 100*time.Minute) // the other 50 minutes are outside the top-N

	if out[0].SharePercent != 30 || out[1].SharePercent != 20 {
		t.Fatalf("shares = %v/%v, want 30/20", out[0].SharePercent, out[1].SharePercent)
	}
}

func TestFromMainStatsNilWhenNothingTracked(t *testing.T) {
	// GetMainStats reports "never tracked anything" as a zero struct. Passing
	// that through would render an activity with no name and a zero streak.
	if got := FromMainStats(models.MainStats{}); got != nil {
		t.Fatalf("FromMainStats(zero) = %+v, want nil", got)
	}
}

func TestFromMainStatsKeepsEmojiAndNameApart(t *testing.T) {
	target := 90
	got := FromMainStats(models.MainStats{
		CurrentActivityID:    5,
		CurrentActivityName:  "Go",
		CurrentActivityEmoji: "🐹",
		TodayTracked:         42 * time.Minute,
		StreakDays:           7,
		TargetMinutes:        &target,
	})

	if got == nil {
		t.Fatal("got nil")
	}
	if got.Name != "Go" {
		t.Fatalf("Name = %q — the emoji must not be glued on", got.Name)
	}
	if got.Emoji != "🐹" {
		t.Fatalf("Emoji = %q", got.Emoji)
	}
	if got.TodaySeconds != 2520 {
		t.Fatalf("TodaySeconds = %d, want 2520", got.TodaySeconds)
	}
	if got.TargetMinutes == nil || *got.TargetMinutes != 90 {
		t.Fatalf("TargetMinutes = %v, want 90", got.TargetMinutes)
	}
}

func TestFromTodayReport(t *testing.T) {
	report := models.ReportTodayStats{
		TotalTracked:  2 * time.Hour,
		TotalSessions: 5,
		TopActivities: []models.ActivityDurationStat{
			{ActivityID: 1, Name: "Go", Emoji: "🐹", Duration: 90 * time.Minute, Sessions: 3},
			{ActivityID: 2, Name: "Sport", Duration: 30 * time.Minute, Sessions: 2},
		},
	}

	got := FromTodayReport(report, 2)
	if got.TotalSeconds != 7200 {
		t.Fatalf("TotalSeconds = %d, want 7200", got.TotalSeconds)
	}
	// Sessions and ActivitiesCount are different numbers from different
	// queries; conflating them is the bug MainStats' field names used to cause.
	if got.Sessions != 5 {
		t.Fatalf("Sessions = %d, want 5", got.Sessions)
	}
	if got.ActivitiesCount != 2 {
		t.Fatalf("ActivitiesCount = %d, want 2", got.ActivitiesCount)
	}
	if got.TopActivities[0].SharePercent != 75 {
		t.Fatalf("first share = %v, want 75", got.TopActivities[0].SharePercent)
	}
}

func TestFromActivityItemsMarksArchived(t *testing.T) {
	target := 60
	items := []models.TrackActivityItem{
		{ID: 1, Name: "Go", Emoji: "🐹", Selected: true, TargetMinutes: &target},
	}

	active := FromActivityItems(items, false)
	if active[0].Archived {
		t.Fatal("active activity marked archived")
	}
	if !active[0].Selected || active[0].TargetMinutes == nil {
		t.Fatalf("fields dropped: %+v", active[0])
	}

	archived := FromActivityItems(items, true)
	if !archived[0].Archived {
		t.Fatal("archived activity not marked")
	}
}

// An empty list must marshal as [] and not null: a client iterating the field
// should not have to guard against nil.
func TestEmptyListsMarshalAsArrays(t *testing.T) {
	raw, err := json.Marshal(ActivitiesResponse{Activities: FromActivityItems(nil, false)})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"activities":[]`) {
		t.Fatalf("expected an empty array, got %s", raw)
	}
}

func TestNewMetaReportsTheUsersZone(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load zone: %v", err)
	}
	meta := NewMeta(loc, time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC))

	if meta.Timezone != "America/New_York" {
		t.Fatalf("Timezone = %q", meta.Timezone)
	}
	// generated_at carries the user's offset, so a client showing it verbatim
	// shows the user's own wall clock.
	if !strings.HasSuffix(string(meta.GeneratedAt), "-04:00") {
		t.Fatalf("GeneratedAt = %q, want the user's offset", meta.GeneratedAt)
	}
}
