//go:build integration

package repo

import (
	"context"
	"testing"
	"time"
)

// TestTimerRepository_UpsertAndGetSettings exercises the real SQL behind
// Activate's "keep the running schedule" logic (see service.timerService.
// Activate): GetSettings must round-trip exactly what UpsertInterval wrote,
// including for a user with no row yet.
func TestTimerRepository_UpsertAndGetSettings(t *testing.T) {
	pool := testPool(t)
	repo := NewTimerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	// No row yet: GetSettings must report "not active", not error.
	_, _, active, err := repo.GetSettings(ctx, userID)
	if err != nil {
		t.Fatalf("get settings (no row): %v", err)
	}
	if active {
		t.Fatalf("get settings (no row): active = true, want false")
	}

	next := time.Now().UTC().Add(30 * time.Minute).Truncate(time.Millisecond)
	if err := repo.UpsertInterval(ctx, userID, 30, next); err != nil {
		t.Fatalf("upsert interval: %v", err)
	}

	gotInterval, gotNext, active, err := repo.GetSettings(ctx, userID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if !active {
		t.Fatalf("get settings: active = false, want true")
	}
	if gotInterval != 30 {
		t.Fatalf("get settings: interval = %d, want 30", gotInterval)
	}
	if !gotNext.Equal(next) {
		t.Fatalf("get settings: next_ping_at = %v, want %v", gotNext, next)
	}

	if err := repo.Disable(ctx, userID); err != nil {
		t.Fatalf("disable: %v", err)
	}
	_, _, active, err = repo.GetSettings(ctx, userID)
	if err != nil {
		t.Fatalf("get settings after disable: %v", err)
	}
	if active {
		t.Fatalf("get settings after disable: active = true, want false")
	}
}
