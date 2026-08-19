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

// LearningRepository persists word-learning collections, words and their
// SRS/push-scheduling state.
type LearningRepository interface {
	// Collections.
	CreateCollection(ctx context.Context, userID int64, name string) (int64, error)
	ListCollections(ctx context.Context, userID int64, archived bool) ([]models.LearningCollectionItem, error)
	GetCollectionName(ctx context.Context, userID, collectionID int64) (string, error)
	ToggleCollectionActive(ctx context.Context, userID, collectionID int64) error
	ArchiveCollection(ctx context.Context, userID, collectionID int64) error
	RestoreCollection(ctx context.Context, userID, collectionID int64) error
	DeleteCollectionForever(ctx context.Context, userID, collectionID int64) error

	// Words.
	AddWords(ctx context.Context, collectionID int64, pairs []models.LearningWordItem) (int, error)
	ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error)
	DeleteWord(ctx context.Context, userID, wordID int64) error
	GetWordForGrading(ctx context.Context, userID, wordID int64) (collectionName, term, translation string, easeFactor float32, intervalDays, repetitions int, err error)
	UpdateWordSchedule(ctx context.Context, wordID int64, easeFactor float32, intervalDays, repetitions int, nextReviewAt time.Time, learned bool) error

	// Review delivery.
	PickDueWord(ctx context.Context, userID int64, now time.Time) (*models.LearningDueWord, error)
	RecordReview(ctx context.Context, wordID, userID int64, correct bool, now time.Time) error

	// Push scheduling — mirrors TimerRepository's shape (migrations/0004).
	UpsertPushInterval(ctx context.Context, userID int64, intervalMin int, nextPushAt time.Time) error
	GetPushSettings(ctx context.Context, userID int64) (intervalMin int, nextPushAt time.Time, enabled bool, err error)
	SetNextPush(ctx context.Context, userID int64, nextPushAt time.Time) error
	DisablePush(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error)

	// Stats.
	CountWords(ctx context.Context, userID int64) (total, dueToday, learned int, err error)
	ListReviewDates(ctx context.Context, userID int64, since time.Time) ([]time.Time, error)
}

type learningRepository struct {
	db *pgxpool.Pool
}

// NewLearningRepository creates repository backed by pgx pool.
func NewLearningRepository(db *pgxpool.Pool) LearningRepository {
	return &learningRepository{db: db}
}

// CreateCollection inserts a new, active, non-archived collection.
func (r *learningRepository) CreateCollection(ctx context.Context, userID int64, name string) (int64, error) {
	q := `
	INSERT INTO learning_collections (user_id, name)
	VALUES ($1, $2)
	RETURNING id;
	`
	var id int64
	err := r.db.QueryRow(ctx, q, userID, name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, models.ErrLearningCollectionExists
		}
		return 0, fmt.Errorf("create collection: %w", err)
	}
	return id, nil
}

// ListCollections returns a user's collections (archived or not) with word
// counts, ordered by creation time.
func (r *learningRepository) ListCollections(ctx context.Context, userID int64, archived bool) ([]models.LearningCollectionItem, error) {
	q := `
	SELECT c.id, c.name, c.is_active, c.is_archived, COUNT(w.id)
	FROM learning_collections c
	LEFT JOIN learning_words w ON w.collection_id = c.id
	WHERE c.user_id = $1 AND c.is_archived = $2
	GROUP BY c.id
	ORDER BY c.created_at;
	`
	rows, err := r.db.Query(ctx, q, userID, archived)
	if err != nil {
		return nil, fmt.Errorf("list collections query: %w", err)
	}
	defer rows.Close()

	out := make([]models.LearningCollectionItem, 0)
	for rows.Next() {
		var item models.LearningCollectionItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Active, &item.IsArchived, &item.WordCount); err != nil {
			return nil, fmt.Errorf("list collections scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list collections rows: %w", err)
	}
	return out, nil
}

// GetCollectionName resolves a collection's name, scoped to its owner.
func (r *learningRepository) GetCollectionName(ctx context.Context, userID, collectionID int64) (string, error) {
	q := `SELECT name FROM learning_collections WHERE id = $1 AND user_id = $2;`
	var name string
	err := r.db.QueryRow(ctx, q, collectionID, userID).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", models.ErrLearningCollectionNotFound
		}
		return "", fmt.Errorf("get collection name: %w", err)
	}
	return name, nil
}

// ToggleCollectionActive flips is_active, scoped to the owning user.
func (r *learningRepository) ToggleCollectionActive(ctx context.Context, userID, collectionID int64) error {
	q := `
	UPDATE learning_collections
	SET is_active = NOT is_active
	WHERE id = $1 AND user_id = $2 AND is_archived = FALSE;
	`
	tag, err := r.db.Exec(ctx, q, collectionID, userID)
	if err != nil {
		return fmt.Errorf("toggle collection active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrLearningCollectionNotFound
	}
	return nil
}

// ArchiveCollection moves a collection to the archive.
func (r *learningRepository) ArchiveCollection(ctx context.Context, userID, collectionID int64) error {
	q := `UPDATE learning_collections SET is_archived = TRUE WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, collectionID, userID)
	if err != nil {
		return fmt.Errorf("archive collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrLearningCollectionNotFound
	}
	return nil
}

// RestoreCollection moves an archived collection back to the active list.
func (r *learningRepository) RestoreCollection(ctx context.Context, userID, collectionID int64) error {
	q := `UPDATE learning_collections SET is_archived = FALSE WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, collectionID, userID)
	if err != nil {
		return fmt.Errorf("restore collection: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrLearningCollectionNotFound
	}
	return nil
}

// DeleteCollectionForever permanently removes a collection and its words
// (ON DELETE CASCADE on learning_words.collection_id).
func (r *learningRepository) DeleteCollectionForever(ctx context.Context, userID, collectionID int64) error {
	q := `DELETE FROM learning_collections WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, collectionID, userID)
	if err != nil {
		return fmt.Errorf("delete collection forever: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrLearningCollectionNotFound
	}
	return nil
}

// AddWords bulk-inserts word pairs into a collection, due immediately (SRS
// defaults from the table's column defaults). Returns the number inserted.
func (r *learningRepository) AddWords(ctx context.Context, collectionID int64, pairs []models.LearningWordItem) (int, error) {
	if len(pairs) == 0 {
		return 0, nil
	}
	batch := &pgx.Batch{}
	for _, p := range pairs {
		batch.Queue(`INSERT INTO learning_words (collection_id, term, translation) VALUES ($1, $2, $3);`,
			collectionID, p.Term, p.Translation)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range pairs {
		if _, err := br.Exec(); err != nil {
			return 0, fmt.Errorf("add words: %w", err)
		}
	}
	return len(pairs), nil
}

// ListWords returns every word in one collection, scoped to its owner.
func (r *learningRepository) ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error) {
	q := `
	SELECT w.id, w.term, w.translation, w.learned, w.next_review_at, w.interval_days, w.repetitions
	FROM learning_words w
	JOIN learning_collections c ON c.id = w.collection_id
	WHERE w.collection_id = $1 AND c.user_id = $2
	ORDER BY w.created_at;
	`
	rows, err := r.db.Query(ctx, q, collectionID, userID)
	if err != nil {
		return nil, fmt.Errorf("list words query: %w", err)
	}
	defer rows.Close()

	out := make([]models.LearningWordItem, 0)
	for rows.Next() {
		var item models.LearningWordItem
		if err := rows.Scan(&item.ID, &item.Term, &item.Translation, &item.Learned, &item.NextReviewAt, &item.IntervalDays, &item.Repetitions); err != nil {
			return nil, fmt.Errorf("list words scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list words rows: %w", err)
	}
	return out, nil
}

// DeleteWord removes one word, scoped to its owner via the collection join.
func (r *learningRepository) DeleteWord(ctx context.Context, userID, wordID int64) error {
	q := `
	DELETE FROM learning_words w
	USING learning_collections c
	WHERE w.id = $1 AND w.collection_id = c.id AND c.user_id = $2;
	`
	tag, err := r.db.Exec(ctx, q, wordID, userID)
	if err != nil {
		return fmt.Errorf("delete word: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrLearningWordNotFound
	}
	return nil
}

// GetWordForGrading loads a word's current SRS state for grading, scoped to
// its owner.
func (r *learningRepository) GetWordForGrading(ctx context.Context, userID, wordID int64) (string, string, string, float32, int, int, error) {
	q := `
	SELECT c.name, w.term, w.translation, w.ease_factor, w.interval_days, w.repetitions
	FROM learning_words w
	JOIN learning_collections c ON c.id = w.collection_id
	WHERE w.id = $1 AND c.user_id = $2;
	`
	var (
		collectionName, term, translation string
		easeFactor                        float32
		intervalDays, repetitions         int
	)
	err := r.db.QueryRow(ctx, q, wordID, userID).Scan(&collectionName, &term, &translation, &easeFactor, &intervalDays, &repetitions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", 0, 0, 0, models.ErrLearningWordNotFound
		}
		return "", "", "", 0, 0, 0, fmt.Errorf("get word for grading: %w", err)
	}
	return collectionName, term, translation, easeFactor, intervalDays, repetitions, nil
}

// UpdateWordSchedule persists new SRS state after grading.
func (r *learningRepository) UpdateWordSchedule(ctx context.Context, wordID int64, easeFactor float32, intervalDays, repetitions int, nextReviewAt time.Time, learned bool) error {
	q := `
	UPDATE learning_words
	SET ease_factor = $2, interval_days = $3, repetitions = $4, next_review_at = $5, learned = $6
	WHERE id = $1;
	`
	_, err := r.db.Exec(ctx, q, wordID, easeFactor, intervalDays, repetitions, nextReviewAt, learned)
	if err != nil {
		return fmt.Errorf("update word schedule: %w", err)
	}
	return nil
}

// PickDueWord returns the single most-overdue due word from the user's
// active, non-archived collections, or nil if none is due.
func (r *learningRepository) PickDueWord(ctx context.Context, userID int64, now time.Time) (*models.LearningDueWord, error) {
	q := `
	SELECT w.id, c.name, w.term, w.translation
	FROM learning_words w
	JOIN learning_collections c ON c.id = w.collection_id
	WHERE c.user_id = $1 AND c.is_active = TRUE AND c.is_archived = FALSE
	  AND w.learned = FALSE AND w.next_review_at <= $2
	ORDER BY w.next_review_at
	LIMIT 1;
	`
	var due models.LearningDueWord
	err := r.db.QueryRow(ctx, q, userID, now).Scan(&due.ID, &due.CollectionName, &due.Term, &due.Translation)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pick due word: %w", err)
	}
	return &due, nil
}

// RecordReview appends one review-history row (used for stats/streak).
func (r *learningRepository) RecordReview(ctx context.Context, wordID, userID int64, correct bool, now time.Time) error {
	q := `INSERT INTO learning_reviews (word_id, user_id, correct, reviewed_at) VALUES ($1, $2, $3, $4);`
	_, err := r.db.Exec(ctx, q, wordID, userID, correct, now)
	if err != nil {
		return fmt.Errorf("record review: %w", err)
	}
	return nil
}

// UpsertPushInterval enables review pushes and saves interval + next push
// timestamp — mirrors TimerRepository.UpsertInterval.
func (r *learningRepository) UpsertPushInterval(ctx context.Context, userID int64, intervalMin int, nextPushAt time.Time) error {
	q := `
	INSERT INTO user_learning_settings (user_id, interval_min, next_push_at, enabled, updated_at)
	VALUES ($1, $2, $3, TRUE, now())
	ON CONFLICT (user_id)
	DO UPDATE SET
		interval_min = EXCLUDED.interval_min,
		next_push_at = EXCLUDED.next_push_at,
		enabled = TRUE,
		updated_at = now();
	`
	if _, err := r.db.Exec(ctx, q, userID, intervalMin, nextPushAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" { // check_violation
			return models.ErrLearningInvalidInterval
		}
		return fmt.Errorf("upsert push interval: %w", err)
	}
	return nil
}

// GetPushSettings returns the raw persisted row for user, if any. enabled is
// false and err is nil when the user has no row yet.
func (r *learningRepository) GetPushSettings(ctx context.Context, userID int64) (int, time.Time, bool, error) {
	q := `SELECT interval_min, next_push_at, enabled FROM user_learning_settings WHERE user_id = $1;`
	var (
		intervalMin int
		nextPushAt  time.Time
		enabled     bool
	)
	err := r.db.QueryRow(ctx, q, userID).Scan(&intervalMin, &nextPushAt, &enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, time.Time{}, false, nil
		}
		return 0, time.Time{}, false, fmt.Errorf("get push settings: %w", err)
	}
	return intervalMin, nextPushAt, enabled, nil
}

// SetNextPush updates next scheduled push time for user.
func (r *learningRepository) SetNextPush(ctx context.Context, userID int64, nextPushAt time.Time) error {
	q := `UPDATE user_learning_settings SET next_push_at = $2, updated_at = now() WHERE user_id = $1;`
	tag, err := r.db.Exec(ctx, q, userID, nextPushAt)
	if err != nil {
		return fmt.Errorf("set next push: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DisablePush turns off review pushes for user.
func (r *learningRepository) DisablePush(ctx context.Context, userID int64) error {
	q := `UPDATE user_learning_settings SET enabled = FALSE, updated_at = now() WHERE user_id = $1;`
	_, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("disable push: %w", err)
	}
	return nil
}

// ListDueUsers returns users whose next_push_at is due.
func (r *learningRepository) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error) {
	q := `
	SELECT uls.user_id, u.tg_user_id, uls.interval_min
	FROM user_learning_settings uls
	JOIN users u ON u.id = uls.user_id
	WHERE uls.enabled = TRUE
	  AND uls.next_push_at IS NOT NULL
	  AND uls.next_push_at <= $1
	ORDER BY uls.next_push_at
	LIMIT $2;
	`
	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due users query: %w", err)
	}
	defer rows.Close()

	out := make([]models.LearningDueUser, 0, limit)
	for rows.Next() {
		var item models.LearningDueUser
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

// CountWords returns total/due-today/learned word counts across all of a
// user's active, non-archived collections.
func (r *learningRepository) CountWords(ctx context.Context, userID int64) (int, int, int, error) {
	q := `
	SELECT
		COUNT(*) FILTER (WHERE TRUE),
		COUNT(*) FILTER (WHERE w.learned = FALSE AND w.next_review_at <= now()),
		COUNT(*) FILTER (WHERE w.learned = TRUE)
	FROM learning_words w
	JOIN learning_collections c ON c.id = w.collection_id
	WHERE c.user_id = $1 AND c.is_archived = FALSE;
	`
	var total, dueToday, learned int
	err := r.db.QueryRow(ctx, q, userID).Scan(&total, &dueToday, &learned)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count words: %w", err)
	}
	return total, dueToday, learned, nil
}

// ListReviewDates returns distinct UTC calendar dates (as midnight
// timestamps) the user answered at least one review on, since the given
// time — used to compute a day streak in the service layer.
func (r *learningRepository) ListReviewDates(ctx context.Context, userID int64, since time.Time) ([]time.Time, error) {
	q := `
	SELECT DISTINCT date_trunc('day', reviewed_at)
	FROM learning_reviews
	WHERE user_id = $1 AND reviewed_at >= $2
	ORDER BY 1 DESC;
	`
	rows, err := r.db.Query(ctx, q, userID, since)
	if err != nil {
		return nil, fmt.Errorf("list review dates query: %w", err)
	}
	defer rows.Close()

	out := make([]time.Time, 0)
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("list review dates scan: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list review dates rows: %w", err)
	}
	return out, nil
}
