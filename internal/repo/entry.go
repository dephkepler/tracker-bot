package repo

import (
	"context"
	"errors"
	"tracker-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryRepository interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetDBIDByTgUserID(ctx context.Context, tgUserID int64) (int64, error)
	Create(ctx context.Context, stats *models.UserInput) (int64, error)
	CountAll(ctx context.Context) (int, error)
	ListPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error)
}
type entryRepository struct {
	db *pgxpool.Pool
}

func NewEntryRepository(db *pgxpool.Pool) EntryRepository {
	return &entryRepository{db: db}
}

func (repo *entryRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	q := `
	SELECT tg_user_id, username, phone_number, email, language, timezone
	FROM users
	WHERE tg_user_id = $1
	`
	var user models.User

	err := repo.db.QueryRow(ctx, q, id).Scan(
		&user.TgUserID,
		&user.UserName,
		&user.PhoneNumber,
		&user.Email,
		&user.Language,
		&user.TimeZone,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (repo *entryRepository) GetDBIDByTgUserID(ctx context.Context, tgUserID int64) (int64, error) {
	q := `
	SELECT id
	FROM users
	WHERE tg_user_id = $1
	`

	var id int64
	err := repo.db.QueryRow(ctx, q, tgUserID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrUserNotFound
		}
		return 0, err
	}

	return id, nil
}

// CountAll returns the total number of registered users.
func (r *entryRepository) CountAll(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM users;`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// ListPage returns one page of users, newest first.
func (r *entryRepository) ListPage(ctx context.Context, limit, offset int) ([]models.AdminUserRow, error) {
	q := `
	SELECT id, tg_user_id, username, created_at
	FROM users
	ORDER BY created_at DESC, id DESC
	LIMIT $1 OFFSET $2;
	`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.AdminUserRow, 0, limit)
	for rows.Next() {
		var u models.AdminUserRow
		if err := rows.Scan(&u.DBID, &u.TgUserID, &u.UserName, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *entryRepository) Create(ctx context.Context, user *models.UserInput) (int64, error) {
	q := `
		INSERT INTO users (tg_user_id, username, phone_number, email, language, timezone)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, q,
		user.TgUserID,
		user.UserName,
		user.PhoneNumber,
		user.Email,
		user.Language,
		user.TimeZone,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
