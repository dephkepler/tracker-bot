//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	"time"
	errlocal "tracker-bot/internal/models"
)

// TestSessionRepository_CreateRetroSession_DedupesSameDueTime covers the
// exact bug reported in prod: a prompt answer button tapped several times
// recorded a full extra session per tap (12 taps -> 3 hours of phantom
// "nothing" time). Repeated calls with the identical endAt (the prompt's
// due time, stable across taps of the same message) must be absorbed into
// a single row, no matter how many times or how far apart they're tapped.
func TestSessionRepository_CreateRetroSession_DedupesSameDueTime(t *testing.T) {
	pool := testPool(t)
	sessionRepo := NewSessionRepository(pool)
	trackRepo := NewTrackerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activity, err := trackRepo.Create(ctx, userID, "Reading", "📖")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	dueAt := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 5; i++ {
		if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 15, "prompt", dueAt); err != nil {
			t.Fatalf("create retro session (tap %d): %v", i, err)
		}
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM activity_sessions WHERE user_id = $1 AND source = 'prompt';`, userID).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("session count after 5 taps on the same due prompt = %d, want 1", count)
	}
}

// TestSessionRepository_CreateRetroSession_DistinctDueTimesBothCredited is
// the other half of the same bug report: answering two different stacked
// prompts together (e.g. one due at 8:34, one at 8:49, both tapped at 9:02)
// must credit both of their real windows, not collapse into one.
func TestSessionRepository_CreateRetroSession_DistinctDueTimesBothCredited(t *testing.T) {
	pool := testPool(t)
	sessionRepo := NewSessionRepository(pool)
	trackRepo := NewTrackerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	activity, err := trackRepo.Create(ctx, userID, "Reading", "📖")
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	first := time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second)
	second := time.Now().UTC().Add(-15 * time.Minute).Truncate(time.Second)
	if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 15, "prompt", first); err != nil {
		t.Fatalf("create retro session (first): %v", err)
	}
	if err := sessionRepo.CreateRetroSession(ctx, userID, activity.ID, 15, "prompt", second); err != nil {
		t.Fatalf("create retro session (second): %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM activity_sessions WHERE user_id = $1 AND source = 'prompt';`, userID).Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 2 {
		t.Fatalf("session count for two distinct due prompts = %d, want 2 (both must be credited)", count)
	}

	var gotFirst, gotSecond time.Time
	rows, err := pool.Query(ctx, `SELECT end_at FROM activity_sessions WHERE user_id = $1 AND source = 'prompt' ORDER BY end_at;`, userID)
	if err != nil {
		t.Fatalf("query end_at: %v", err)
	}
	defer rows.Close()
	var ends []time.Time
	for rows.Next() {
		var e time.Time
		if err := rows.Scan(&e); err != nil {
			t.Fatalf("scan end_at: %v", err)
		}
		ends = append(ends, e)
	}
	if len(ends) != 2 {
		t.Fatalf("end_at rows = %v, want 2 distinct values", ends)
	}
	gotFirst, gotSecond = ends[0], ends[1]
	if !gotFirst.Equal(first) || !gotSecond.Equal(second) {
		t.Fatalf("end_at values = (%v, %v), want (%v, %v)", gotFirst, gotSecond, first, second)
	}
}

// TestSessionRepository_CreateRetroSession_InvalidActivity checks a
// nonexistent/foreign/archived activity still surfaces ErrActivityNotFound.
func TestSessionRepository_CreateRetroSession_InvalidActivity(t *testing.T) {
	pool := testPool(t)
	sessionRepo := NewSessionRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if err := sessionRepo.CreateRetroSession(ctx, userID, 999999, 15, "prompt", time.Now().UTC()); !errors.Is(err, errlocal.ErrActivityNotFound) {
		t.Fatalf("create retro session for nonexistent activity: got %v, want ErrActivityNotFound", err)
	}
}
