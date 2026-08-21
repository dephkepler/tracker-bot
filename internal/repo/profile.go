package repo

import (
	"context"
	"errors"
	"tracker-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository interface {
	Create(ctx context.Context, stats *models.ProfileStats) error
	GetByID(ctx context.Context, id int64) (*models.ProfileStats, error)
	Update(ctx context.Context, id int64, stats *models.ProfileStats) error
	Delete(ctx context.Context, id int64) error
}
type profileRepository struct {
	db *pgxpool.Pool
}

func NewProfileRepository(db *pgxpool.Pool) ProfileRepository {
	return &profileRepository{db: db}
}

func (repo *profileRepository) Create(ctx context.Context, stats *models.ProfileStats) error {
	q := `
		INSERT INTO users (tg_user_id, username, phone_number, email, language, timezone)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := repo.db.Exec(ctx, q, stats.TgUserID, stats.UserName, stats.PhoneNumber, stats.Email, stats.Language, stats.TimeZone)
	if err != nil {
		return err
	}

	return nil
}

func (repo *profileRepository) GetByID(ctx context.Context, id int64) (*models.ProfileStats, error) {
	q := `
	SELECT tg_user_id, username, phone_number, email, language, timezone
	FROM users
	WHERE tg_user_id = $1
	`
	var profile models.ProfileStats

	err := repo.db.QueryRow(ctx, q, id).Scan(
		&profile.TgUserID,
		&profile.UserName,
		&profile.PhoneNumber,
		&profile.Email,
		&profile.Language,
		&profile.TimeZone,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	return &profile, nil
}

func (repo *profileRepository) Update(ctx context.Context, id int64, stats *models.ProfileStats) error {
	// id is the Telegram user id (matches GetByID/Create) — using the DB id here caused silent "user not found" for every user.
	q := `
		UPDATE users
		SET language = $2, timezone = $3
		WHERE tg_user_id = $1
	`

	res, err := repo.db.Exec(ctx, q, id, stats.Language, stats.TimeZone)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (repo *profileRepository) Delete(ctx context.Context, id int64) error {
	q := `
		DELETE FROM users
		WHERE id = $1
	`

	res, err := repo.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}
