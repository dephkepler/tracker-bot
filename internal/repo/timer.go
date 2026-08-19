package repo

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tracker-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TimerRepository works with persisted timer settings.
type TimerRepository interface {
	UpsertInterval(ctx context.Context, userID int64, intervalMin int, nextPingAt time.Time) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error)
	SetNextPing(ctx context.Context, userID int64, nextPingAt time.Time) error
	GetInterval(ctx context.Context, userID int64) (int, error)
	GetSettings(ctx context.Context, userID int64) (intervalMin int, nextPingAt time.Time, enabled bool, err error)
	Disable(ctx context.Context, userID int64) error
}

type timerRepository struct {
	db *pgxpool.Pool
}

// NewTimerRepository creates repository backed by pgx pool.
func NewTimerRepository(db *pgxpool.Pool) TimerRepository {
	return &timerRepository{db: db}
}

// UpsertInterval enables timer and saves interval + next ping timestamp.
func (r *timerRepository) UpsertInterval(ctx context.Context, userID int64, intervalMin int, nextPingAt time.Time) error {
	q := `
	INSERT INTO user_timer_settings (user_id, interval_min, next_ping_at, enabled, updated_at)
	VALUES ($1, $2, $3, TRUE, now())
	ON CONFLICT (user_id)
	DO UPDATE SET
		interval_min = EXCLUDED.interval_min,
		next_ping_at = EXCLUDED.next_ping_at,
		enabled = TRUE,
		updated_at = now();
	`
	if _, err := r.db.Exec(ctx, q, userID, intervalMin, nextPingAt); err != nil {
		return fmt.Errorf("upsert interval: %w", err)
	}
	return nil
}

// ListDueUsers returns users whose next_ping_at is due.
func (r *timerRepository) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error) {
	q := `
	SELECT uts.user_id, u.tg_user_id, uts.interval_min
	FROM user_timer_settings uts
	JOIN users u ON u.id = uts.user_id
	WHERE uts.enabled = TRUE
	  AND uts.next_ping_at IS NOT NULL
	  AND uts.next_ping_at <= $1
	ORDER BY uts.next_ping_at
	LIMIT $2;
	`

	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due users query: %w", err)
	}
	defer rows.Close()

	out := make([]models.TimerDueUser, 0, limit)
	for rows.Next() {
		var item models.TimerDueUser
		if err := rows.Scan(&item.DBUserID, &item.TgUserID, &item.IntervalMin); err != nil {
			return nil, fmt.Errorf("list due users scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due users rows: %w", err)
	}
	return out, nil
}

// SetNextPing updates next scheduled ping time for user.
func (r *timerRepository) SetNextPing(ctx context.Context, userID int64, nextPingAt time.Time) error {
	q := `
	UPDATE user_timer_settings
	SET next_ping_at = $2, updated_at = now()
	WHERE user_id = $1;
	`
	tag, err := r.db.Exec(ctx, q, userID, nextPingAt)
	if err != nil {
		return fmt.Errorf("set next ping: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetInterval returns active timer interval in minutes.
func (r *timerRepository) GetInterval(ctx context.Context, userID int64) (int, error) {
	q := `
	SELECT interval_min
	FROM user_timer_settings
	WHERE user_id = $1 AND enabled = TRUE;
	`
	var interval int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&interval); err != nil {
		return 0, err
	}
	return interval, nil
}

// GetSettings returns the raw persisted timer row for user, if any.
// enabled is false and err is nil when the user has no timer row yet.
func (r *timerRepository) GetSettings(ctx context.Context, userID int64) (int, time.Time, bool, error) {
	q := `
	SELECT interval_min, next_ping_at, enabled
	FROM user_timer_settings
	WHERE user_id = $1;
	`
	var (
		intervalMin int
		nextPingAt  time.Time
		enabled     bool
	)
	err := r.db.QueryRow(ctx, q, userID).Scan(&intervalMin, &nextPingAt, &enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, false, nil
		}
		return 0, time.Time{}, false, fmt.Errorf("get settings: %w", err)
	}
	return intervalMin, nextPingAt, enabled, nil
}

// Disable turns off timer for user.
func (r *timerRepository) Disable(ctx context.Context, userID int64) error {
	q := `
	UPDATE user_timer_settings
	SET enabled = FALSE, updated_at = now()
	WHERE user_id = $1;
	`
	_, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("disable timer: %w", err)
	}
	return nil
}
