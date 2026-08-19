//go:build integration

package repo

import (
	"context"
	"testing"
	"time"
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
	if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 30, "test"); err != nil {
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
