package repo

import (
	"context"
	"errors"
	"fmt"
	errlocal "tracker-bot/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CustomTimerRepository persists user-defined timer intervals (in addition
// to the built-in 15/30 min choices).
type CustomTimerRepository interface {
	// Create adds a custom interval for user. Idempotent: adding an
	// interval that already exists for this user is a no-op, not an error.
	Create(ctx context.Context, userID int64, intervalMin int) error
	ListByUser(ctx context.Context, userID int64) ([]int, error)
	Delete(ctx context.Context, userID int64, intervalMin int) error
	Count(ctx context.Context, userID int64) (int, error)
}

type customTimerRepository struct {
	db *pgxpool.Pool
}

// NewCustomTimerRepository creates repository backed by pgx pool.
func NewCustomTimerRepository(db *pgxpool.Pool) CustomTimerRepository {
	return &customTimerRepository{db: db}
}

// Create inserts a custom interval for user, ignoring duplicates.
func (r *customTimerRepository) Create(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("create custom timer: invalid args")
	}

	q := `
	INSERT INTO user_custom_timers (user_id, interval_min)
	VALUES ($1, $2)
	ON CONFLICT (user_id, interval_min) DO NOTHING;
	`
	if _, err := r.db.Exec(ctx, q, userID, intervalMin); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" { // check_violation
			return errlocal.ErrCustomTimerInvalidInterval
		}
		return fmt.Errorf("create custom timer: %w", err)
	}
	return nil
}

// ListByUser returns custom intervals for user, ascending.
func (r *customTimerRepository) ListByUser(ctx context.Context, userID int64) ([]int, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("list custom timers: invalid userID")
	}

	q := `
	SELECT interval_min
	FROM user_custom_timers
	WHERE user_id = $1
	ORDER BY interval_min;
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list custom timers query: %w", err)
	}
	defer rows.Close()

	out := make([]int, 0, 8)
	for rows.Next() {
		var m int
		if err := rows.Scan(&m); err != nil {
			return nil, fmt.Errorf("list custom timers scan: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list custom timers rows: %w", err)
	}
	return out, nil
}

// Delete removes one custom interval for user.
func (r *customTimerRepository) Delete(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("delete custom timer: invalid args")
	}

	q := `
	DELETE FROM user_custom_timers
	WHERE user_id = $1 AND interval_min = $2;
	`
	tag, err := r.db.Exec(ctx, q, userID, intervalMin)
	if err != nil {
		return fmt.Errorf("delete custom timer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errlocal.ErrCustomTimerNotFound
	}
	return nil
}

// Count returns how many custom intervals user already has.
func (r *customTimerRepository) Count(ctx context.Context, userID int64) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("count custom timers: invalid userID")
	}

	q := `SELECT count(*) FROM user_custom_timers WHERE user_id = $1;`
	var n int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count custom timers: %w", err)
	}
	return n, nil
}
