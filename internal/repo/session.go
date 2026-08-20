package repo

import (
	"context"
	"fmt"
	errlocal "tracker-bot/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository stores tracked activity sessions.
type SessionRepository interface {
	// CreateRetroSession saves one session ending "now" with duration intervalMin.
	CreateRetroSession(ctx context.Context, userID, activityID int64, intervalMin int, source string) error
}

type sessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates session repository backed by pgx pool.
func NewSessionRepository(db *pgxpool.Pool) SessionRepository {
	return &sessionRepository{db: db}
}

// CreateRetroSession writes a backfilled session only for user's active activity.
func (r *sessionRepository) CreateRetroSession(ctx context.Context, userID, activityID int64, intervalMin int, source string) error {
	if userID <= 0 || activityID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("create retro session: invalid input")
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

	q := `
	INSERT INTO activity_sessions (user_id, activity_id, start_at, end_at, planned_min, source)
	SELECT $1, $2, now() - make_interval(mins => $3), now(), $3, $4
	-- Guards against a burst of duplicate inserts from the same prompt round
	-- (e.g. the user tapping an answer button several times before Telegram
	-- deletes it, or a client retry) — a real second answer is always at
	-- least one full interval later, far outside this window. Silently
	-- absorbed rather than erroring: the first tap already recorded it.
	WHERE NOT EXISTS (
		SELECT 1
		FROM activity_sessions
		WHERE user_id = $1 AND source = $4 AND end_at > now() - interval '30 seconds'
	);
	`
	if _, err := r.db.Exec(ctx, q, userID, activityID, intervalMin, source); err != nil {
		return fmt.Errorf("create retro session exec: %w", err)
	}
	return nil
}
