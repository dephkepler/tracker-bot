//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// TestChallengeRepository_CreatePrePopulatesDays checks a new challenge gets
// one challenge_days row per day in its range, all pending, plus the
// unique-name and max-length constraints.
func TestChallengeRepository_CreatePrePopulatesDays(t *testing.T) {
	pool := testPool(t)
	repo := NewChallengeRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 9) // 10 days total

	id, err := repo.Create(ctx, userID, "10 days", start, end)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	days, err := repo.ListDays(ctx, userID, id)
	if err != nil {
		t.Fatalf("list days: %v", err)
	}
	if len(days) != 10 {
		t.Fatalf("days = %d, want 10", len(days))
	}
	for _, d := range days {
		if d.Status != errlocal.ChallengeDayPending {
			t.Fatalf("day %v status = %v, want pending", d.Date, d.Status)
		}
	}

	if _, err := repo.Create(ctx, userID, "10 days", start, end); !errors.Is(err, errlocal.ErrChallengeExists) {
		t.Fatalf("duplicate name: got %v, want ErrChallengeExists", err)
	}

	if _, err := repo.Create(ctx, userID, "Way too long", start, start.AddDate(0, 0, 150)); !errors.Is(err, errlocal.ErrChallengeInvalidRange) {
		t.Fatalf("150-day range: got %v, want ErrChallengeInvalidRange", err)
	}
}

// TestChallengeRepository_MarkDayAndProgress checks marking days updates
// GetDayStatus and ListChallenges' done/skipped counts.
func TestChallengeRepository_MarkDayAndProgress(t *testing.T) {
	pool := testPool(t)
	repo := NewChallengeRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 2) // 3 days
	id, err := repo.Create(ctx, userID, "3 days", start, end)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.MarkDay(ctx, userID, id, start, errlocal.ChallengeDayDone); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	if err := repo.MarkDay(ctx, userID, id, start.AddDate(0, 0, 1), errlocal.ChallengeDaySkipped); err != nil {
		t.Fatalf("mark skipped: %v", err)
	}

	status, err := repo.GetDayStatus(ctx, userID, id, start)
	if err != nil || status != errlocal.ChallengeDayDone {
		t.Fatalf("get day status = (%v, %v), want (done, nil)", status, err)
	}

	items, err := repo.ListChallenges(ctx, userID, false)
	if err != nil {
		t.Fatalf("list challenges: %v", err)
	}
	if len(items) != 1 || items[0].TotalDays != 3 || items[0].DoneDays != 1 || items[0].SkippedDays != 1 {
		t.Fatalf("challenge progress = %+v, want TotalDays=3 DoneDays=1 SkippedDays=1", items)
	}

	if err := repo.MarkDay(ctx, userID, id, start.AddDate(0, 0, 30), errlocal.ChallengeDayDone); !errors.Is(err, errlocal.ErrChallengeDayNotFound) {
		t.Fatalf("mark day outside range: got %v, want ErrChallengeDayNotFound", err)
	}
}

// TestChallengeRepository_ArchiveRestoreDelete mirrors the learning
// collection lifecycle test.
func TestChallengeRepository_ArchiveRestoreDelete(t *testing.T) {
	pool := testPool(t)
	repo := NewChallengeRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 4)
	id, err := repo.Create(ctx, userID, "5 days", start, end)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.UpsertPushSchedule(ctx, id, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("upsert push schedule: %v", err)
	}

	if err := repo.ArchiveChallenge(ctx, userID, id); err != nil {
		t.Fatalf("archive: %v", err)
	}
	active, _ := repo.ListChallenges(ctx, userID, false)
	if len(active) != 0 {
		t.Fatalf("active list after archive = %+v, want empty", active)
	}
	// Archiving must also clear the push schedule.
	due, err := repo.ListDueChallenges(ctx, time.Now().UTC().Add(2*time.Hour), 100)
	if err != nil {
		t.Fatalf("list due challenges: %v", err)
	}
	for _, d := range due {
		if d.ChallengeID == id {
			t.Fatalf("archived challenge %d still due for push", id)
		}
	}

	if err := repo.RestoreChallenge(ctx, userID, id); err != nil {
		t.Fatalf("restore: %v", err)
	}
	active, _ = repo.ListChallenges(ctx, userID, false)
	if len(active) != 1 {
		t.Fatalf("active list after restore = %+v, want one entry", active)
	}

	if err := repo.DeleteChallengeForever(ctx, userID, id); err != nil {
		t.Fatalf("delete forever: %v", err)
	}
	active, _ = repo.ListChallenges(ctx, userID, false)
	if len(active) != 0 {
		t.Fatalf("active list after delete forever = %+v, want empty", active)
	}
	if _, err := repo.ListDays(ctx, userID, id); err != nil {
		t.Fatalf("list days after delete: %v", err)
	}
}

// TestChallengeRepository_ListDueChallenges checks the due-list respects
// next_push_at and excludes archived challenges.
func TestChallengeRepository_ListDueChallenges(t *testing.T) {
	pool := testPool(t)
	repo := NewChallengeRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 9)
	id, err := repo.Create(ctx, userID, "due test", start, end)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	now := time.Now().UTC()
	if err := repo.UpsertPushSchedule(ctx, id, now.Add(-time.Minute)); err != nil {
		t.Fatalf("upsert push schedule: %v", err)
	}

	due, err := repo.ListDueChallenges(ctx, now, 100)
	if err != nil {
		t.Fatalf("list due challenges: %v", err)
	}
	found := false
	for _, d := range due {
		if d.ChallengeID == id {
			found = true
			if !d.StartDate.Equal(start) || !d.EndDate.Equal(end) {
				t.Fatalf("due entry dates = (%v, %v), want (%v, %v)", d.StartDate, d.EndDate, start, end)
			}
		}
	}
	if !found {
		t.Fatalf("list due challenges = %+v, want to include challenge %d", due, id)
	}

	if err := repo.ClearPushSchedule(ctx, id); err != nil {
		t.Fatalf("clear push schedule: %v", err)
	}
	due, _ = repo.ListDueChallenges(ctx, now, 100)
	for _, d := range due {
		if d.ChallengeID == id {
			t.Fatalf("challenge %d still due after ClearPushSchedule", id)
		}
	}
}
