package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	errlocal "tracker-bot/internal/models"
	"tracker-bot/pkg/apptime"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Activity struct {
	ID         int64
	UserID     int64
	Name       string
	Emoji      string
	IsArchived bool
	CreatedAt  time.Time
}
type TrackerRepository interface {
	Create(ctx context.Context, userID int64, name, emoji string) (Activity, error)
	ListActive(ctx context.Context, userID int64) ([]Activity, error)
	ListArchived(ctx context.Context, userID int64) ([]Activity, error)
	SelectedListActive(ctx context.Context, userID int64) ([]int64, error)
	ToggleSelectedActive(ctx context.Context, userID, activityID int64) error
	DeleteSelected(ctx context.Context, userID int64) (int64, error)
	ArchiveSelected(ctx context.Context, userID int64) (int64, error)
	RestoreArchived(ctx context.Context, userID, activityID int64) error
	DeleteArchivedForever(ctx context.Context, userID, activityID int64) error
	GetTodayStats(ctx context.Context, userID int64, tzName string) (time.Duration, int, error)
	GetTodayActivities(ctx context.Context, userID int64, tzName string) ([]Activity, []time.Duration, []int, error)
	GetPeriodActivities(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64) ([]Activity, []time.Duration, []int, time.Duration, int, error)
	GetPeriodMonthlyTotals(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, tzName string) ([]time.Time, []time.Duration, error)
	GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location, tzName string) (map[int]time.Duration, error)
	GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, tzName string) ([]time.Time, []time.Duration, error)
	GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, tzName string) ([]errlocal.HourActivityDuration, error)
	GetLastTrackedActiveActivity(ctx context.Context, userID int64) (Activity, bool, error)
	GetTodayDurationByActivity(ctx context.Context, userID, activityID int64, tzName string) (time.Duration, error)
	GetTrackedDaysDescByActivity(ctx context.Context, userID, activityID int64, tzName string) ([]time.Time, error)
	GetTodayTrackedActivitiesCount(ctx context.Context, userID int64, tzName string) (int, error)
}
type trackRepository struct {
	db *pgxpool.Pool
}

func NewTrackerRepository(db *pgxpool.Pool) TrackerRepository {
	return &trackRepository{db: db}
}

func (r *trackRepository) Create(ctx context.Context, userID int64, name, emoji string) (Activity, error) {
	if userID <= 0 {
		return Activity{}, fmt.Errorf("create activity: invalid userID")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Activity{}, fmt.Errorf("create activity: empty name")
	}

	emoji = strings.TrimSpace(emoji)

	q := `
	INSERT INTO activities (user_id, name, emoji)
	VALUES ($1, $2, $3)
	RETURNING id, user_id, name, emoji, is_archived, created_at;
	`

	var a Activity
	err := r.db.QueryRow(ctx, q, userID, name, emoji).Scan(
		&a.ID,
		&a.UserID,
		&a.Name,
		&a.Emoji,
		&a.IsArchived,
		&a.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// 23505 is PostgreSQL unique_violation.
			if pgErr.Code == "23505" {
				return Activity{}, errlocal.ErrActivityExists
			}
		}
		return Activity{}, fmt.Errorf("create activity: %w", err)
	}
	return a, nil
}

func (r *trackRepository) ListActive(ctx context.Context, userID int64) ([]Activity, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("list active: invalid userID")
	}

	q := `
	SELECT id, user_id, name, emoji, is_archived, created_at
	FROM activities
	WHERE user_id = $1 AND is_archived = false
	ORDER BY lower(name), id;
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list active query: %w", err)
	}
	defer rows.Close()

	out := make([]Activity, 0, 16)
	for rows.Next() {
		var a Activity
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Name, &a.Emoji, &a.IsArchived, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("list active scan: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active rows: %w", err)
	}
	return out, nil
}

func (r *trackRepository) SelectedListActive(ctx context.Context, userID int64) ([]int64, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("selected list: invalid userID")
	}

	q := `
	SELECT activity_id
	FROM user_selected_activities
	WHERE user_id = $1
	ORDER BY activity_id;
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("selected list query: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("selected list scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("selected list rows: %w", err)
	}
	return ids, nil
}

func (r *trackRepository) ListArchived(ctx context.Context, userID int64) ([]Activity, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("list archived: invalid userID")
	}

	q := `
	SELECT id, user_id, name, emoji, is_archived, created_at
	FROM activities
	WHERE user_id = $1 AND is_archived = true
	ORDER BY lower(name), id;
	`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list archived query: %w", err)
	}
	defer rows.Close()

	out := make([]Activity, 0, 16)
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Emoji, &a.IsArchived, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("list archived scan: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list archived rows: %w", err)
	}
	return out, nil
}

// ToggleSelectedActive toggles one activity in user_selected_activities.
// A transaction keeps ownership check and toggle operation atomic.
func (r *trackRepository) ToggleSelectedActive(ctx context.Context, userID, activityID int64) error {
	if userID <= 0 || activityID <= 0 {
		return fmt.Errorf("toggle selected: invalid ids")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("toggle selected begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1) Ensure the activity belongs to this user and is active.
	ownQ := `
	SELECT EXISTS(
		SELECT 1
		FROM activities
		WHERE id = $1 AND user_id = $2 AND is_archived = false
	);`
	var owned bool
	if err := tx.QueryRow(ctx, ownQ, activityID, userID).Scan(&owned); err != nil {
		return fmt.Errorf("toggle selected ownership: %w", err)
	}
	if !owned {
		return errlocal.ErrActivityNotFound
	}

	// 2) Try to unselect first.
	delQ := `
	DELETE FROM user_selected_activities
	WHERE user_id = $1 AND activity_id = $2;`
	tag, err := tx.Exec(ctx, delQ, userID, activityID)
	if err != nil {
		return fmt.Errorf("toggle selected delete: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return tx.Commit(ctx)
	}

	// 3) If nothing was deleted, select it.
	insQ := `
	INSERT INTO user_selected_activities(user_id, activity_id)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING;`
	if _, err := tx.Exec(ctx, insQ, userID, activityID); err != nil {
		return fmt.Errorf("toggle selected insert: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *trackRepository) DeleteSelected(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("delete selected: invalid userID")
	}

	q := `
	DELETE FROM activities a
	USING user_selected_activities s
	WHERE a.id = s.activity_id
	  AND a.user_id = $1
	  AND s.user_id = $1;
	`

	tag, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return 0, fmt.Errorf("delete selected exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *trackRepository) ArchiveSelected(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("archive selected: invalid userID")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("archive selected begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	updateQ := `
	UPDATE activities a
	SET is_archived = TRUE
	FROM user_selected_activities s
	WHERE a.id = s.activity_id
	  AND a.user_id = $1
	  AND s.user_id = $1;
	`
	tag, err := tx.Exec(ctx, updateQ, userID)
	if err != nil {
		return 0, fmt.Errorf("archive selected update: %w", err)
	}

	cleanupQ := `DELETE FROM user_selected_activities WHERE user_id = $1;`
	if _, err := tx.Exec(ctx, cleanupQ, userID); err != nil {
		return 0, fmt.Errorf("archive selected cleanup: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("archive selected commit: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *trackRepository) RestoreArchived(ctx context.Context, userID, activityID int64) error {
	if userID <= 0 || activityID <= 0 {
		return fmt.Errorf("restore archived: invalid input")
	}
	q := `
	UPDATE activities
	SET is_archived = FALSE
	WHERE id = $1 AND user_id = $2 AND is_archived = TRUE;
	`
	tag, err := r.db.Exec(ctx, q, activityID, userID)
	if err != nil {
		return fmt.Errorf("restore archived exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errlocal.ErrActivityNotFound
	}
	return nil
}

func (r *trackRepository) DeleteArchivedForever(ctx context.Context, userID, activityID int64) error {
	if userID <= 0 || activityID <= 0 {
		return fmt.Errorf("delete archived forever: invalid input")
	}
	q := `
	DELETE FROM activities
	WHERE id = $1 AND user_id = $2 AND is_archived = TRUE;
	`
	tag, err := r.db.Exec(ctx, q, activityID, userID)
	if err != nil {
		return fmt.Errorf("delete archived forever exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errlocal.ErrActivityNotFound
	}
	return nil
}

func (r *trackRepository) GetTodayStats(ctx context.Context, userID int64, tzName string) (time.Duration, int, error) {
	if userID <= 0 {
		return 0, 0, fmt.Errorf("today stats: invalid userID")
	}
	q := `
	SELECT
		COALESCE(SUM(end_at - start_at), interval '0'),
		COUNT(*)
	FROM activity_sessions
	WHERE user_id = $1
	  AND end_at IS NOT NULL
	  AND (start_at AT TIME ZONE $2) >= date_trunc('day', now() AT TIME ZONE $2)
	  AND (start_at AT TIME ZONE $2) <  date_trunc('day', now() AT TIME ZONE $2) + interval '1 day';
	`
	var total time.Duration
	var sessions int
	if err := r.db.QueryRow(ctx, q, userID, tzName).Scan(&total, &sessions); err != nil {
		return 0, 0, fmt.Errorf("today stats query: %w", err)
	}
	return total, sessions, nil
}

func (r *trackRepository) GetTodayActivities(ctx context.Context, userID int64, tzName string) ([]Activity, []time.Duration, []int, error) {
	if userID <= 0 {
		return nil, nil, nil, fmt.Errorf("today activities: invalid userID")
	}
	q := `
	SELECT
		a.id, a.user_id, a.name, COALESCE(a.emoji, ''), a.is_archived, a.created_at,
		COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur,
		COUNT(*) AS sessions
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND (s.start_at AT TIME ZONE $2) >= date_trunc('day', now() AT TIME ZONE $2)
	  AND (s.start_at AT TIME ZONE $2) <  date_trunc('day', now() AT TIME ZONE $2) + interval '1 day'
	GROUP BY a.id, a.user_id, a.name, a.emoji, a.is_archived, a.created_at
	ORDER BY total_dur DESC;
	`
	rows, err := r.db.Query(ctx, q, userID, tzName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("today activities query: %w", err)
	}
	defer rows.Close()

	activities := make([]Activity, 0, 16)
	durations := make([]time.Duration, 0, 16)
	sessions := make([]int, 0, 16)
	for rows.Next() {
		var a Activity
		var dur time.Duration
		var cnt int
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Emoji, &a.IsArchived, &a.CreatedAt, &dur, &cnt); err != nil {
			return nil, nil, nil, fmt.Errorf("today activities scan: %w", err)
		}
		activities = append(activities, a)
		durations = append(durations, dur)
		sessions = append(sessions, cnt)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("today activities rows: %w", err)
	}
	return activities, durations, sessions, nil
}

func (r *trackRepository) GetPeriodActivities(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64) ([]Activity, []time.Duration, []int, time.Duration, int, error) {
	if userID <= 0 {
		return nil, nil, nil, 0, 0, fmt.Errorf("period activities: invalid userID")
	}
	if from.After(to) {
		return nil, nil, nil, 0, 0, fmt.Errorf("period activities: invalid range")
	}
	if len(activityIDs) == 0 {
		return nil, nil, nil, 0, 0, nil
	}

	q := `
	SELECT
		a.id, a.user_id, a.name, COALESCE(a.emoji, ''), a.is_archived, a.created_at,
		COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur,
		COUNT(*) AS sessions
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND s.start_at >= $2
	  AND s.start_at < $3
	  AND s.activity_id = ANY($4)
	GROUP BY a.id, a.user_id, a.name, a.emoji, a.is_archived, a.created_at
	ORDER BY total_dur DESC;
	`

	rows, err := r.db.Query(ctx, q, userID, from.UTC(), to.UTC(), activityIDs)
	if err != nil {
		return nil, nil, nil, 0, 0, fmt.Errorf("period activities query: %w", err)
	}
	defer rows.Close()

	activities := make([]Activity, 0, len(activityIDs))
	durations := make([]time.Duration, 0, len(activityIDs))
	sessions := make([]int, 0, len(activityIDs))
	var total time.Duration
	var totalSessions int

	for rows.Next() {
		var a Activity
		var dur time.Duration
		var cnt int
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Emoji, &a.IsArchived, &a.CreatedAt, &dur, &cnt); err != nil {
			return nil, nil, nil, 0, 0, fmt.Errorf("period activities scan: %w", err)
		}
		activities = append(activities, a)
		durations = append(durations, dur)
		sessions = append(sessions, cnt)
		total += dur
		totalSessions += cnt
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, 0, 0, fmt.Errorf("period activities rows: %w", err)
	}

	return activities, durations, sessions, total, totalSessions, nil
}

func (r *trackRepository) GetPeriodMonthlyTotals(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, tzName string) ([]time.Time, []time.Duration, error) {
	if userID <= 0 || len(activityIDs) == 0 {
		return nil, nil, nil
	}
	q := `
	SELECT date_trunc('month', s.start_at AT TIME ZONE $5) AS month_start,
	       COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND s.start_at >= $2
	  AND s.start_at < $3
	  AND s.activity_id = ANY($4)
	GROUP BY month_start
	ORDER BY month_start;
	`
	rows, err := r.db.Query(ctx, q, userID, from.UTC(), to.UTC(), activityIDs, tzName)
	if err != nil {
		return nil, nil, fmt.Errorf("period monthly query: %w", err)
	}
	defer rows.Close()
	months := make([]time.Time, 0, 16)
	durs := make([]time.Duration, 0, 16)
	for rows.Next() {
		var m time.Time
		var d time.Duration
		if err := rows.Scan(&m, &d); err != nil {
			return nil, nil, fmt.Errorf("period monthly scan: %w", err)
		}
		months = append(months, m)
		durs = append(durs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("period monthly rows: %w", err)
	}
	return months, durs, nil
}

func (r *trackRepository) GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location, tzName string) (map[int]time.Duration, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("month daily totals: invalid userID")
	}
	if len(activityIDs) == 0 {
		return map[int]time.Duration{}, nil
	}
	if loc == nil {
		loc = apptime.Location
	}
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, loc)
	next := first.AddDate(0, 1, 0)
	q := `
	SELECT EXTRACT(DAY FROM s.start_at AT TIME ZONE $5)::int AS day_num,
	       COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND s.start_at >= $2
	  AND s.start_at < $3
	  AND s.activity_id = ANY($4)
	GROUP BY day_num
	ORDER BY day_num;
	`
	rows, err := r.db.Query(ctx, q, userID, first, next, activityIDs, tzName)
	if err != nil {
		return nil, fmt.Errorf("month daily totals query: %w", err)
	}
	defer rows.Close()
	out := make(map[int]time.Duration)
	for rows.Next() {
		var day int
		var dur time.Duration
		if err := rows.Scan(&day, &dur); err != nil {
			return nil, fmt.Errorf("month daily totals scan: %w", err)
		}
		out[day] = dur
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("month daily totals rows: %w", err)
	}
	return out, nil
}

func (r *trackRepository) GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, tzName string) ([]time.Time, []time.Duration, error) {
	if userID <= 0 || len(activityIDs) == 0 {
		return nil, nil, nil
	}
	if granularity != "month" && granularity != "day" && granularity != "hour" {
		return nil, nil, fmt.Errorf("invalid granularity")
	}
	q := fmt.Sprintf(`
	SELECT date_trunc('%s', s.start_at AT TIME ZONE $5) AS bucket_start,
	       COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND s.start_at >= $2
	  AND s.start_at < $3
	  AND s.activity_id = ANY($4)
	GROUP BY bucket_start
	ORDER BY bucket_start;
	`, granularity)
	rows, err := r.db.Query(ctx, q, userID, from.UTC(), to.UTC(), activityIDs, tzName)
	if err != nil {
		return nil, nil, fmt.Errorf("period buckets query: %w", err)
	}
	defer rows.Close()

	buckets := make([]time.Time, 0, 64)
	durs := make([]time.Duration, 0, 64)
	for rows.Next() {
		var b time.Time
		var d time.Duration
		if err := rows.Scan(&b, &d); err != nil {
			return nil, nil, fmt.Errorf("period buckets scan: %w", err)
		}
		buckets = append(buckets, b)
		durs = append(durs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("period buckets rows: %w", err)
	}
	return buckets, durs, nil
}

// GetHourlyBucketsByActivity returns per-hour totals broken down by
// activity, so a "By hours" report can show *what* was tracked in each
// hour instead of just a total minute count. Rows are ordered by hour, then
// by duration descending within that hour.
func (r *trackRepository) GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, tzName string) ([]errlocal.HourActivityDuration, error) {
	if userID <= 0 || len(activityIDs) == 0 {
		return nil, nil
	}
	q := `
	SELECT date_trunc('hour', s.start_at AT TIME ZONE $5) AS bucket_start,
	       a.name, COALESCE(a.emoji, ''),
	       COALESCE(SUM(s.end_at - s.start_at), interval '0') AS total_dur
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND s.start_at >= $2
	  AND s.start_at < $3
	  AND s.activity_id = ANY($4)
	GROUP BY bucket_start, a.id, a.name, a.emoji
	ORDER BY bucket_start, total_dur DESC;
	`
	rows, err := r.db.Query(ctx, q, userID, from.UTC(), to.UTC(), activityIDs, tzName)
	if err != nil {
		return nil, fmt.Errorf("hourly buckets by activity query: %w", err)
	}
	defer rows.Close()

	out := make([]errlocal.HourActivityDuration, 0, 64)
	for rows.Next() {
		var item errlocal.HourActivityDuration
		if err := rows.Scan(&item.BucketStart, &item.Name, &item.Emoji, &item.Duration); err != nil {
			return nil, fmt.Errorf("hourly buckets by activity scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hourly buckets by activity rows: %w", err)
	}
	return out, nil
}

func (r *trackRepository) GetLastTrackedActiveActivity(ctx context.Context, userID int64) (Activity, bool, error) {
	if userID <= 0 {
		return Activity{}, false, fmt.Errorf("last tracked active activity: invalid userID")
	}
	q := `
	SELECT a.id, a.user_id, a.name, COALESCE(a.emoji, ''), a.is_archived, a.created_at
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	ORDER BY COALESCE(s.end_at, s.start_at) DESC
	LIMIT 1;
	`
	var a Activity
	err := r.db.QueryRow(ctx, q, userID).Scan(&a.ID, &a.UserID, &a.Name, &a.Emoji, &a.IsArchived, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Activity{}, false, nil
	}
	if err != nil {
		return Activity{}, false, fmt.Errorf("last tracked active activity query: %w", err)
	}
	return a, true, nil
}

func (r *trackRepository) GetTodayDurationByActivity(ctx context.Context, userID, activityID int64, tzName string) (time.Duration, error) {
	if userID <= 0 || activityID <= 0 {
		return 0, fmt.Errorf("today duration by activity: invalid input")
	}
	q := `
	SELECT COALESCE(SUM(s.end_at - s.start_at), interval '0')
	FROM activity_sessions s
	WHERE s.user_id = $1
	  AND s.activity_id = $2
	  AND s.end_at IS NOT NULL
	  AND (s.start_at AT TIME ZONE $3) >= date_trunc('day', now() AT TIME ZONE $3)
	  AND (s.start_at AT TIME ZONE $3) <  date_trunc('day', now() AT TIME ZONE $3) + interval '1 day';
	`
	var total time.Duration
	if err := r.db.QueryRow(ctx, q, userID, activityID, tzName).Scan(&total); err != nil {
		return 0, fmt.Errorf("today duration by activity query: %w", err)
	}
	return total, nil
}

func (r *trackRepository) GetTrackedDaysDescByActivity(ctx context.Context, userID, activityID int64, tzName string) ([]time.Time, error) {
	if userID <= 0 || activityID <= 0 {
		return nil, fmt.Errorf("tracked days by activity: invalid input")
	}
	q := `
	SELECT DISTINCT date_trunc('day', s.start_at AT TIME ZONE $3)::timestamp
	FROM activity_sessions s
	WHERE s.user_id = $1
	  AND s.activity_id = $2
	  AND s.end_at IS NOT NULL
	ORDER BY 1 DESC;
	`
	rows, err := r.db.Query(ctx, q, userID, activityID, tzName)
	if err != nil {
		return nil, fmt.Errorf("tracked days by activity query: %w", err)
	}
	defer rows.Close()

	out := make([]time.Time, 0, 64)
	for rows.Next() {
		var day time.Time
		if err := rows.Scan(&day); err != nil {
			return nil, fmt.Errorf("tracked days by activity scan: %w", err)
		}
		out = append(out, day.UTC())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tracked days by activity rows: %w", err)
	}
	return out, nil
}

func (r *trackRepository) GetTodayTrackedActivitiesCount(ctx context.Context, userID int64, tzName string) (int, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("today tracked activities count: invalid userID")
	}
	q := `
	SELECT COUNT(DISTINCT s.activity_id)
	FROM activity_sessions s
	JOIN activities a ON a.id = s.activity_id
	WHERE s.user_id = $1
	  AND a.is_archived = FALSE
	  AND s.end_at IS NOT NULL
	  AND (s.start_at AT TIME ZONE $2) >= date_trunc('day', now() AT TIME ZONE $2)
	  AND (s.start_at AT TIME ZONE $2) <  date_trunc('day', now() AT TIME ZONE $2) + interval '1 day';
	`
	var count int
	if err := r.db.QueryRow(ctx, q, userID, tzName).Scan(&count); err != nil {
		return 0, fmt.Errorf("today tracked activities count query: %w", err)
	}
	return count, nil
}
