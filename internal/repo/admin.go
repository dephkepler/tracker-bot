package repo

import (
	"context"
	"errors"
	"tracker-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepository serves cross-domain reporting queries for the admin
// screens (bot-wide overview, per-user drill-down) — deliberately separate
// from EntryRepository, which only ever touches the users table.
type AdminRepository interface {
	GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error)
	GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error)
}

type adminRepository struct {
	db *pgxpool.Pool
}

// NewAdminRepository creates repository backed by pgx pool.
func NewAdminRepository(db *pgxpool.Pool) AdminRepository {
	return &adminRepository{db: db}
}

// GetOverviewStats aggregates bot-wide usage numbers with a handful of
// simple counts — admin-only, low traffic, not worth a fancier single query.
func (r *adminRepository) GetOverviewStats(ctx context.Context) (models.AdminOverviewStats, error) {
	var s models.AdminOverviewStats

	counts := []struct {
		q   string
		dst *int
	}{
		{`SELECT count(*) FROM users;`, &s.TotalUsers},
		{`SELECT count(*) FROM user_timer_settings WHERE enabled = TRUE;`, &s.ActiveTrackTimers},
		{`SELECT count(*) FROM user_learning_settings WHERE enabled = TRUE;`, &s.ActiveReviewPushes},
		{`SELECT count(*) FROM activities WHERE is_archived = FALSE;`, &s.TotalActivities},
		{`SELECT count(*) FROM learning_collections WHERE is_archived = FALSE;`, &s.TotalCollections},
		{`SELECT count(*) FROM learning_words;`, &s.TotalLearningWords},
	}
	for _, c := range counts {
		if err := r.db.QueryRow(ctx, c.q).Scan(c.dst); err != nil {
			return models.AdminOverviewStats{}, err
		}
	}
	return s, nil
}

// GetUserDetail loads one user's profile fields plus cross-domain usage
// counts, for the admin's per-user drill-down.
func (r *adminRepository) GetUserDetail(ctx context.Context, dbUserID int64) (models.AdminUserDetail, error) {
	var d models.AdminUserDetail
	q := `
	SELECT
		u.id, u.tg_user_id, u.username, u.language, u.timezone, u.created_at,
		(SELECT count(*) FROM activities a WHERE a.user_id = u.id AND a.is_archived = FALSE),
		(SELECT count(*) FROM learning_collections c WHERE c.user_id = u.id AND c.is_archived = FALSE),
		(SELECT count(*) FROM learning_words w JOIN learning_collections c ON c.id = w.collection_id WHERE c.user_id = u.id),
		COALESCE((SELECT enabled FROM user_timer_settings ts WHERE ts.user_id = u.id), FALSE),
		COALESCE((SELECT enabled FROM user_learning_settings ls WHERE ls.user_id = u.id), FALSE)
	FROM users u
	WHERE u.id = $1;
	`
	err := r.db.QueryRow(ctx, q, dbUserID).Scan(
		&d.DBID, &d.TgUserID, &d.UserName, &d.Language, &d.TimeZone, &d.CreatedAt,
		&d.ActivitiesCount, &d.CollectionsCount, &d.LearningWords,
		&d.TrackTimerActive, &d.ReviewsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AdminUserDetail{}, models.ErrUserNotFound
		}
		return models.AdminUserDetail{}, err
	}
	return d, nil
}
