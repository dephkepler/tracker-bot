package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UIStateRepository interface {
	GetScreen(ctx context.Context, userID int64) (string, error)
	SetScreen(ctx context.Context, userID int64, screen string) error
}

type uiStateRepository struct {
	db *pgxpool.Pool
}

func NewUIStateRepository(db *pgxpool.Pool) UIStateRepository {
	return &uiStateRepository{db: db}
}

// Returns "" with no error when the user has no saved screen yet.
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
