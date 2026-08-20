//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	errlocal "tracker-bot/internal/models"
)

// TestSessionRepository_CreateRetroSession_DedupesRapidDuplicates covers the
// exact bug reported in prod: a prompt answer button tapped several times
// within seconds recorded a full extra session per tap (12 taps -> 3 hours
// of phantom "nothing" time). A second call within the cooldown window must
// be silently absorbed, not create a second row.
func TestSessionRepository_CreateRetroSession_DedupesRapidDuplicates(t *testing.T) {
	pool := testPool(t)
	sessionRepo := NewSessionRepository(pool)
	trackRepo := NewTrackerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activity, err := trackRepo.Create(ctx, userID, "Reading", "📖")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 15, "prompt"); err != nil {
			t.Fatalf("create retro session (tap %d): %v", i, err)
		}
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM activity_sessions WHERE user_id = $1 AND source = 'prompt';`, userID).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("session count after 5 rapid taps = %d, want 1 (duplicates within the cooldown window must be absorbed)", count)
	}
}

// TestSessionRepository_CreateRetroSession_InvalidActivity checks a
// nonexistent/foreign/archived activity still surfaces ErrActivityNotFound.
func TestSessionRepository_CreateRetroSession_InvalidActivity(t *testing.T) {
	pool := testPool(t)
	sessionRepo := NewSessionRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if err := sessionRepo.CreateRetroSession(ctx, userID, 999999, 15, "prompt"); !errors.Is(err, errlocal.ErrActivityNotFound) {
		t.Fatalf("create retro session for nonexistent activity: got %v, want ErrActivityNotFound", err)
	}
}
