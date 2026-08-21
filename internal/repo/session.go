package repo

import (
	"context"
	"fmt"
	"time"
	errlocal "tracker-bot/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository interface {
	// idempotent: re-answering the same (user, activity, source, endAt) is silently absorbed, not double-credited
	CreateRetroSession(ctx context.Context, userID, activityID int64, intervalMin int, source string, endAt time.Time) error
}

type sessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) CreateRetroSession(ctx context.Context, userID, activityID int64, intervalMin int, source string, endAt time.Time) error {
	if userID <= 0 || activityID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("create retro session: invalid input")
	}
	if endAt.IsZero() {
		endAt = time.Now().UTC()
	}

	var activityExists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM activities WHERE id = $1 AND user_id = $2 AND is_archived = FALSE);`,
		activityID, userID,
	).Scan(&activityExists)
	if err != nil {
		return fmt.Errorf("create retro session check activity: %w", err)
	}
	if !activityExists {
		return errlocal.ErrActivityNotFound
	}

	startAt := endAt.Add(-time.Duration(intervalMin) * time.Minute)
	q := `
	INSERT INTO activity_sessions (user_id, activity_id, start_at, end_at, planned_min, source)
	SELECT $1, $2, $3::timestamptz, $4::timestamptz, $5, $6
	WHERE NOT EXISTS (
		SELECT 1
		FROM activity_sessions
		WHERE user_id = $1 AND activity_id = $2 AND source = $6 AND end_at = $4::timestamptz
	);
	`
	if _, err := r.db.Exec(ctx, q, userID, activityID, startAt, endAt, intervalMin, source); err != nil {
		return fmt.Errorf("create retro session exec: %w", err)
	}
	return nil
}
