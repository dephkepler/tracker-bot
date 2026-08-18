package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UIStateRepository persists which "screen" a user is currently on, so
// navigation survives a bot restart instead of resetting to zero for
// everyone mid-flow.
type UIStateRepository interface {
	GetScreen(ctx context.Context, userID int64) (string, error)
	SetScreen(ctx context.Context, userID int64, screen string) error
}

type uiStateRepository struct {
	db *pgxpool.Pool
}

// NewUIStateRepository creates repository backed by pgx pool.
func NewUIStateRepository(db *pgxpool.Pool) UIStateRepository {
	return &uiStateRepository{db: db}
}

// GetScreen returns the saved screen for user, or "" if none saved yet.
func (r *uiStateRepository) GetScreen(ctx context.Context, userID int64) (string, error) {
	q := `SELECT screen FROM user_ui_state WHERE user_id = $1;`
	var screen string
	if err := r.db.QueryRow(ctx, q, userID).Scan(&screen); err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get screen: %w", err)
	}
	return screen, nil
}

// SetScreen upserts the current screen for user.
func (r *uiStateRepository) SetScreen(ctx context.Context, userID int64, screen string) error {
	q := `
	INSERT INTO user_ui_state (user_id, screen, updated_at)
	VALUES ($1, $2, now())
	ON CONFLICT (user_id)
	DO UPDATE SET screen = EXCLUDED.screen, updated_at = now();
	`
	if _, err := r.db.Exec(ctx, q, userID, screen); err != nil {
		return fmt.Errorf("set screen: %w", err)
	}
	return nil
}
