package repo

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tracker-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChallengeRepository interface {
	Create(ctx context.Context, userID int64, name string, startDate, endDate time.Time) (int64, error)
	ListChallenges(ctx context.Context, userID int64, archived bool) ([]models.ChallengeItem, error)
	GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error)
	ArchiveChallenge(ctx context.Context, userID, challengeID int64) error
	RestoreChallenge(ctx context.Context, userID, challengeID int64) error
	DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error

	ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error)
	MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, status models.ChallengeDayStatus) error
	GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error)

	// Unlike TimerRepository, "next" here is always "tomorrow, same wall-clock time" — caller computes it from the user's timezone.
	UpsertPushSchedule(ctx context.Context, challengeID int64, nextPushAt time.Time) error
	ClearPushSchedule(ctx context.Context, challengeID int64) error
	ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error)
}

type challengeRepository struct {
	db *pgxpool.Pool
}

func NewChallengeRepository(db *pgxpool.Pool) ChallengeRepository {
	return &challengeRepository{db: db}
}

// Inserts into challenges, then challenge_days for the full range — two calls, not one transaction.
func (r *challengeRepository) Create(ctx context.Context, userID int64, name string, startDate, endDate time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO challenges (user_id, name, start_date, end_date) VALUES ($1, $2, $3, $4) RETURNING id;`,
		userID, name, startDate, endDate,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return 0, models.ErrChallengeExists
			case "23514":
				return 0, models.ErrChallengeInvalidRange
			}
		}
		return 0, fmt.Errorf("create challenge: %w", err)
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO challenge_days (challenge_id, day_date)
		 SELECT $1, d::date FROM generate_series($2::date, $3::date, interval '1 day') AS d;`,
		id, startDate, endDate,
	)
	if err != nil {
		return 0, fmt.Errorf("create challenge days: %w", err)
	}
	return id, nil
}

func (r *challengeRepository) ListChallenges(ctx context.Context, userID int64, archived bool) ([]models.ChallengeItem, error) {
	q := `
	SELECT c.id, c.name, c.start_date, c.end_date, c.is_archived,
		COUNT(cd.id),
		COUNT(cd.id) FILTER (WHERE cd.status = 'done'),
		COUNT(cd.id) FILTER (WHERE cd.status = 'skipped')
	FROM challenges c
	LEFT JOIN challenge_days cd ON cd.challenge_id = c.id
	WHERE c.user_id = $1 AND c.is_archived = $2
	GROUP BY c.id
	ORDER BY c.created_at;
	`
	rows, err := r.db.Query(ctx, q, userID, archived)
	if err != nil {
		return nil, fmt.Errorf("list challenges query: %w", err)
	}
	defer rows.Close()

	out := make([]models.ChallengeItem, 0)
	for rows.Next() {
		var item models.ChallengeItem
		if err := rows.Scan(&item.ID, &item.Name, &item.StartDate, &item.EndDate, &item.IsArchived, &item.TotalDays, &item.DoneDays, &item.SkippedDays); err != nil {
			return nil, fmt.Errorf("list challenges scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list challenges rows: %w", err)
	}
	return out, nil
}

func (r *challengeRepository) GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error) {
	q := `
	SELECT c.id, c.name, c.start_date, c.end_date, c.is_archived,
		COUNT(cd.id),
		COUNT(cd.id) FILTER (WHERE cd.status = 'done'),
		COUNT(cd.id) FILTER (WHERE cd.status = 'skipped')
	FROM challenges c
	LEFT JOIN challenge_days cd ON cd.challenge_id = c.id
	WHERE c.id = $1 AND c.user_id = $2
	GROUP BY c.id;
	`
	var item models.ChallengeItem
	err := r.db.QueryRow(ctx, q, challengeID, userID).Scan(&item.ID, &item.Name, &item.StartDate, &item.EndDate, &item.IsArchived, &item.TotalDays, &item.DoneDays, &item.SkippedDays)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ChallengeItem{}, models.ErrChallengeNotFound
		}
		return models.ChallengeItem{}, fmt.Errorf("get challenge: %w", err)
	}
	return item, nil
}

func (r *challengeRepository) ArchiveChallenge(ctx context.Context, userID, challengeID int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE challenges SET is_archived = TRUE, next_push_at = NULL WHERE id = $1 AND user_id = $2;`, challengeID, userID)
	if err != nil {
		return fmt.Errorf("archive challenge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrChallengeNotFound
	}
	return nil
}

// Caller must re-schedule the push separately if the challenge is still in range.
func (r *challengeRepository) RestoreChallenge(ctx context.Context, userID, challengeID int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE challenges SET is_archived = FALSE WHERE id = $1 AND user_id = $2;`, challengeID, userID)
	if err != nil {
		return fmt.Errorf("restore challenge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrChallengeNotFound
	}
	return nil
}

func (r *challengeRepository) DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM challenges WHERE id = $1 AND user_id = $2;`, challengeID, userID)
	if err != nil {
		return fmt.Errorf("delete challenge forever: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrChallengeNotFound
	}
	return nil
}

func (r *challengeRepository) ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error) {
	q := `
	SELECT cd.day_date, cd.status
	FROM challenge_days cd
	JOIN challenges c ON c.id = cd.challenge_id
	WHERE c.id = $1 AND c.user_id = $2
	ORDER BY cd.day_date;
	`
	rows, err := r.db.Query(ctx, q, challengeID, userID)
	if err != nil {
		return nil, fmt.Errorf("list challenge days query: %w", err)
	}
	defer rows.Close()

	out := make([]models.ChallengeDay, 0)
	for rows.Next() {
		var d models.ChallengeDay
		var status string
		if err := rows.Scan(&d.Date, &status); err != nil {
			return nil, fmt.Errorf("list challenge days scan: %w", err)
		}
		d.Status = models.ChallengeDayStatus(status)
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list challenge days rows: %w", err)
	}
	return out, nil
}

func (r *challengeRepository) MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, status models.ChallengeDayStatus) error {
	q := `
	UPDATE challenge_days cd
	SET status = $4, marked_at = now()
	FROM challenges c
	WHERE cd.challenge_id = c.id AND c.id = $1 AND c.user_id = $2 AND cd.day_date = $3;
	`
	tag, err := r.db.Exec(ctx, q, challengeID, userID, day, string(status))
	if err != nil {
		return fmt.Errorf("mark challenge day: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrChallengeDayNotFound
	}
	return nil
}

func (r *challengeRepository) GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error) {
	q := `
	SELECT cd.status
	FROM challenge_days cd
	JOIN challenges c ON c.id = cd.challenge_id
	WHERE c.id = $1 AND c.user_id = $2 AND cd.day_date = $3;
	`
	var status string
	err := r.db.QueryRow(ctx, q, challengeID, userID, day).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", models.ErrChallengeDayNotFound
		}
		return "", fmt.Errorf("get challenge day status: %w", err)
	}
	return models.ChallengeDayStatus(status), nil
}

func (r *challengeRepository) UpsertPushSchedule(ctx context.Context, challengeID int64, nextPushAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE challenges SET next_push_at = $2 WHERE id = $1;`, challengeID, nextPushAt)
	if err != nil {
		return fmt.Errorf("upsert challenge push schedule: %w", err)
	}
	return nil
}

func (r *challengeRepository) ClearPushSchedule(ctx context.Context, challengeID int64) error {
	_, err := r.db.Exec(ctx, `UPDATE challenges SET next_push_at = NULL WHERE id = $1;`, challengeID)
	if err != nil {
		return fmt.Errorf("clear challenge push schedule: %w", err)
	}
	return nil
}

func (r *challengeRepository) ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error) {
	q := `
	SELECT c.id, c.user_id, u.tg_user_id, c.name, c.start_date, c.end_date
	FROM challenges c
	JOIN users u ON u.id = c.user_id
	WHERE c.is_archived = FALSE
	  AND c.next_push_at IS NOT NULL
	  AND c.next_push_at <= $1
	ORDER BY c.next_push_at
	LIMIT $2;
	`
	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due challenges query: %w", err)
	}
	defer rows.Close()

	out := make([]models.ChallengeDueUser, 0, limit)
	for rows.Next() {
		var item models.ChallengeDueUser
		if err := rows.Scan(&item.ChallengeID, &item.DBUserID, &item.TgUserID, &item.ChallengeName, &item.StartDate, &item.EndDate); err != nil {
			return nil, fmt.Errorf("list due challenges scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due challenges rows: %w", err)
	}
	return out, nil
}
