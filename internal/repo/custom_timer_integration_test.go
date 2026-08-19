//go:build integration

package repo

import (
	"context"
	"errors"
	"testing"
	errlocal "tracker-bot/internal/models"
)

// TestCustomTimerRepository_CreateListDelete exercises the real SQL against
// a live Postgres: insert, idempotent re-insert, ordering, and delete.
func TestCustomTimerRepository_CreateListDelete(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomTimerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if err := repo.Create(ctx, userID, 30); err != nil {
		t.Fatalf("create 30: %v", err)
	}
	if err := repo.Create(ctx, userID, 5); err != nil {
		t.Fatalf("create 5: %v", err)
	}
	// Re-adding the same interval must be a no-op, not a unique-violation error.
	if err := repo.Create(ctx, userID, 5); err != nil {
		t.Fatalf("create 5 again (should be idempotent): %v", err)
	}

	got, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []int{5, 30}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("list = %v, want %v", got, want)
	}

	n, err := repo.Count(ctx, userID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}

	if err := repo.Delete(ctx, userID, 5); err != nil {
		t.Fatalf("delete 5: %v", err)
	}
	got, err = repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(got) != 1 || got[0] != 30 {
		t.Fatalf("list after delete = %v, want [30]", got)
	}
}

// TestCustomTimerRepository_Delete_NotFound checks deleting a non-existent
// interval surfaces the domain "not found" error instead of silently
// succeeding.
func TestCustomTimerRepository_Delete_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomTimerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	err := repo.Delete(ctx, userID, 45)
	if !errors.Is(err, errlocal.ErrCustomTimerNotFound) {
		t.Fatalf("delete missing interval: got %v, want ErrCustomTimerNotFound", err)
	}
}

// TestCustomTimerRepository_RangeConstraint checks the DB-level CHECK
// constraint (chk_custom_interval_min_range) is actually enforced, as a
// defensive backstop behind the service-layer validation.
func TestCustomTimerRepository_RangeConstraint(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomTimerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	err := repo.Create(ctx, userID, 361)
	if !errors.Is(err, errlocal.ErrCustomTimerInvalidInterval) {
		t.Fatalf("create out-of-range interval: got %v, want ErrCustomTimerInvalidInterval", err)
	}
}

// TestCustomTimerRepository_CascadeOnUserDelete checks custom timers are
// removed automatically when the owning user is deleted.
func TestCustomTimerRepository_CascadeOnUserDelete(t *testing.T) {
	pool := testPool(t)
	repo := NewCustomTimerRepository(pool)
	ctx := context.Background()
	userID := testUser(t, pool)

	if err := repo.Create(ctx, userID, 20); err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1;`, userID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	n, err := repo.Count(ctx, userID)
	if err != nil {
		t.Fatalf("count after user delete: %v", err)
	}
	if n != 0 {
		t.Fatalf("count after user delete = %d, want 0 (cascade should have removed it)", n)
	}
}
