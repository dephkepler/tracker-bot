package service

import (
	"context"
	"errors"
	"testing"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// naiveDay builds a value shaped exactly like what the repository hands to
// calcStreakDays. repo.GetTrackedDaysDescByActivity selects
// `date_trunc('day', s.start_at AT TIME ZONE $3)::timestamp` — a
// `timestamp without time zone`, i.e. a local wall clock with no offset — and
// pgx returns it with time.UTC attached. So "2026-08-21 in the user's zone"
// arrives as 2026-08-21T00:00:00Z regardless of what that zone is.
//
// Getting this wrong in a test would hide the very bug these cases exist for,
// which is why the shape lives in one helper.
func naiveDay(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return loc
}

func TestCalcStreakDaysAcrossTimezones(t *testing.T) {
	// Three consecutive tracked days ending today, in every direction of UTC
	// offset. The streak must be 3 everywhere: the day list is already in the
	// user's own zone, so the answer cannot depend on the offset.
	//
	// This is the regression guard for a real bug — the implementation used to
	// call .In(loc) on the naive values above, which reinterpreted a local wall
	// clock as a UTC instant and shifted it. Positive offsets survived by
	// accident (Warsaw +2 lands back on the same date); every zone west of UTC
	// got keys one day early, so today never matched and the streak read 0.
	days := []time.Time{
		naiveDay(2026, time.August, 21),
		naiveDay(2026, time.August, 20),
		naiveDay(2026, time.August, 19),
	}

	cases := []struct {
		zone string
		want int
	}{
		{"Europe/Warsaw", 3},    // +02:00 — passed even before the fix
		{"America/New_York", 3}, // −04:00 — read 0 before the fix
		{"Pacific/Auckland", 3}, // +12:00
		{"UTC", 3},
		{"Asia/Kolkata", 3}, // +05:30 — a non-whole-hour offset
	}

	for _, tc := range cases {
		t.Run(tc.zone, func(t *testing.T) {
			loc := mustLoadLocation(t, tc.zone)
			now := time.Date(2026, time.August, 21, 12, 0, 0, 0, loc)

			if got := calcStreakDays(days, now, loc); got != tc.want {
				t.Fatalf("calcStreakDays in %s = %d, want %d", tc.zone, got, tc.want)
			}
		})
	}
}

func TestCalcStreakDaysNearMidnight(t *testing.T) {
	// The hour of "now" must not matter — only which local day it falls on.
	// A user checking their streak at 00:30 and at 23:30 sees the same number.
	loc := mustLoadLocation(t, "America/New_York")
	days := []time.Time{
		naiveDay(2026, time.August, 21),
		naiveDay(2026, time.August, 20),
	}

	for _, hour := range []int{0, 1, 12, 23} {
		now := time.Date(2026, time.August, 21, hour, 30, 0, 0, loc)
		if got := calcStreakDays(days, now, loc); got != 2 {
			t.Fatalf("calcStreakDays at %02d:30 = %d, want 2", hour, got)
		}
	}
}

func TestCalcStreakDaysStopsAtGap(t *testing.T) {
	loc := mustLoadLocation(t, "America/New_York")
	days := []time.Time{
		naiveDay(2026, time.August, 21),
		naiveDay(2026, time.August, 19), // the 20th is missing
		naiveDay(2026, time.August, 18),
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, loc)

	if got := calcStreakDays(days, now, loc); got != 1 {
		t.Fatalf("calcStreakDays = %d, want 1 — the streak stops at the gap", got)
	}
}

func TestCalcStreakDaysRequiresToday(t *testing.T) {
	// Existing behaviour, pinned rather than changed: the streak is counted
	// back from today, so an unbroken run that ended yesterday reads 0. The bot
	// has always shown it this way.
	loc := mustLoadLocation(t, "Europe/Warsaw")
	days := []time.Time{
		naiveDay(2026, time.August, 20),
		naiveDay(2026, time.August, 19),
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, loc)

	if got := calcStreakDays(days, now, loc); got != 0 {
		t.Fatalf("calcStreakDays = %d, want 0 when today is missing", got)
	}
}

func TestCalcStreakDaysEmpty(t *testing.T) {
	loc := mustLoadLocation(t, "Europe/Warsaw")
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, loc)

	if got := calcStreakDays(nil, now, loc); got != 0 {
		t.Fatalf("calcStreakDays(nil) = %d, want 0", got)
	}
}

func TestCalcStreakDaysNilLocation(t *testing.T) {
	// resolveLoc falls back to apptime.Location, so a nil zone must not panic
	// and must still count the run. Callers pass ctx.Location, which the
	// dispatcher guarantees, but GetMainStats is reachable with nil in tests.
	days := []time.Time{naiveDay(2026, time.August, 21)}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)

	if got := calcStreakDays(days, now, nil); got != 1 {
		t.Fatalf("calcStreakDays with nil loc = %d, want 1", got)
	}
}

func TestCalcStreakDaysAcrossDSTFallBack(t *testing.T) {
	// Warsaw leaves DST on 2026-10-25 (+02:00 → +01:00). Day arithmetic must
	// walk calendar days, not subtract 24h, or the 25th gets visited twice and
	// the 26th never.
	loc := mustLoadLocation(t, "Europe/Warsaw")
	days := []time.Time{
		naiveDay(2026, time.October, 26),
		naiveDay(2026, time.October, 25),
		naiveDay(2026, time.October, 24),
	}
	now := time.Date(2026, time.October, 26, 12, 0, 0, 0, loc)

	if got := calcStreakDays(days, now, loc); got != 3 {
		t.Fatalf("calcStreakDays across the DST change = %d, want 3", got)
	}
}

// --- SetActivityTarget ------------------------------------------------

// fakeTargetRepo embeds a nil repo.TrackerRepository and overrides only
// SetActivityTarget — every other method would panic if called, which is
// fine since these tests only exercise the target-setting path.
type fakeTargetRepo struct {
	repo.TrackerRepository
	gotUserID, gotActivityID int64
	gotMinutes               int
	called                   bool
}

func (f *fakeTargetRepo) SetActivityTarget(_ context.Context, userID, activityID int64, minutes int) error {
	f.called = true
	f.gotUserID, f.gotActivityID, f.gotMinutes = userID, activityID, minutes
	return nil
}

func TestSetActivityTarget_RejectsOutOfRange(t *testing.T) {
	// Invalid input is rejected before the repo is ever consulted, so a nil
	// repo is safe here.
	svc := NewTrackerService(nil)
	ctx := context.Background()

	for _, minutes := range []int{0, -5, 1441} {
		if err := svc.SetActivityTarget(ctx, 1, 1, minutes); !errors.Is(err, models.ErrActivityTargetInvalid) {
			t.Errorf("SetActivityTarget(%d): got %v, want ErrActivityTargetInvalid", minutes, err)
		}
	}
}

func TestSetActivityTarget_AcceptsBoundsAndPassesThrough(t *testing.T) {
	ctx := context.Background()

	for _, minutes := range []int{models.MinActivityTargetMinutes, 90, models.MaxActivityTargetMinutes} {
		fake := &fakeTargetRepo{}
		svc := NewTrackerService(fake)
		if err := svc.SetActivityTarget(ctx, 7, 42, minutes); err != nil {
			t.Fatalf("SetActivityTarget(%d): got %v, want nil", minutes, err)
		}
		if !fake.called {
			t.Fatalf("SetActivityTarget(%d): repo was never called", minutes)
		}
		if fake.gotUserID != 7 || fake.gotActivityID != 42 || fake.gotMinutes != minutes {
			t.Errorf("SetActivityTarget(%d): repo got (user=%d, activity=%d, minutes=%d), want (7, 42, %d)",
				minutes, fake.gotUserID, fake.gotActivityID, fake.gotMinutes, minutes)
		}
	}
}
