//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// TestTrackerRepository_GetTrackedDaysInRange checks a completed session
// makes its calendar day show up within a matching range, and that the
// range bounds are actually respected.
func TestTrackerRepository_GetTrackedDaysInRange(t *testing.T) {
	pool := testPool(t)
	trackRepo := NewTrackerRepository(pool)
	sessionRepo := NewSessionRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activity, err := trackRepo.Create(ctx, userID, "Reading", "📖")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}
	if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 30, "test", time.Now().UTC()); err != nil {
		t.Fatalf("create retro session: %v", err)
	}

	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)

	days, err := trackRepo.GetTrackedDaysInRange(ctx, userID, from, to, "UTC")
	if err != nil {
		t.Fatalf("get tracked days in range: %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("tracked days = %+v, want exactly today", days)
	}

	// A range entirely in the past must not include today's session.
	pastFrom := now.Add(-72 * time.Hour)
	pastTo := now.Add(-48 * time.Hour)
	days, err = trackRepo.GetTrackedDaysInRange(ctx, userID, pastFrom, pastTo, "UTC")
	if err != nil {
		t.Fatalf("get tracked days in past range: %v", err)
	}
	if len(days) != 0 {
		t.Fatalf("tracked days in past range = %+v, want empty (out of bounds)", days)
	}
}

// TestTrackerRepository_SetActivityTarget covers the daily-target column:
// set/read-back via ListActive and GetLastTrackedActiveActivity, the CHECK
// constraint's range enforcement, and cross-user ownership.
func TestTrackerRepository_SetActivityTarget(t *testing.T) {
	pool := testPool(t)
	trackRepo := NewTrackerRepository(pool)
	sessionRepo := NewSessionRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activity, err := trackRepo.Create(ctx, userID, "Reading", "📖")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	items, err := trackRepo.ListActive(ctx, userID)
	if err != nil || len(items) != 1 || items[0].TargetMinutes != nil {
		t.Fatalf("list active before set = (%+v, %v), want one activity with nil target", items, err)
	}

	if err := trackRepo.SetActivityTarget(ctx, userID, activity.ID, 90); err != nil {
		t.Fatalf("set activity target: %v", err)
	}

	items, err = trackRepo.ListActive(ctx, userID)
	if err != nil || len(items) != 1 || items[0].TargetMinutes == nil || *items[0].TargetMinutes != 90 {
		t.Fatalf("list active after set = (%+v, %v), want target 90", items, err)
	}

	// GetLastTrackedActiveActivity must also surface the target — it's the
	// query GetMainStats relies on for the progress bar.
	if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 30, "test", time.Now().UTC()); err != nil {
		t.Fatalf("create retro session: %v", err)
	}
	last, ok, err := trackRepo.GetLastTrackedActiveActivity(ctx, userID)
	if err != nil || !ok || last.TargetMinutes == nil || *last.TargetMinutes != 90 {
		t.Fatalf("get last tracked active activity = (%+v, %v, %v), want target 90", last, ok, err)
	}

	// The CHECK constraint's range is surfaced as a typed error, not a raw
	// pg error, same convention as the other domains' interval checks.
	if err := trackRepo.SetActivityTarget(ctx, userID, activity.ID, 0); !errors.Is(err, errlocal.ErrActivityTargetInvalid) {
		t.Fatalf("set target 0: got %v, want ErrActivityTargetInvalid", err)
	}
	if err := trackRepo.SetActivityTarget(ctx, userID, activity.ID, 1441); !errors.Is(err, errlocal.ErrActivityTargetInvalid) {
		t.Fatalf("set target 1441: got %v, want ErrActivityTargetInvalid", err)
	}

	// Another user must not be able to set this activity's target.
	otherUser := testUser(t, pool)
	if err := trackRepo.SetActivityTarget(ctx, otherUser, activity.ID, 60); !errors.Is(err, errlocal.ErrActivityNotFound) {
		t.Fatalf("cross-user set target: got %v, want ErrActivityNotFound", err)
	}
}
