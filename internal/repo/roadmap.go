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

// RoadmapRepository persists learning roadmaps (one per technology), their
// freeform checklist cards, and the digest-push schedule. Mirrors
// LearningRepository's shape minus everything SRS-related — a card is just
// done or pending, there's no per-card review schedule.
type RoadmapRepository interface {
	// Roadmaps.
	CreateRoadmap(ctx context.Context, userID int64, name string) (int64, error)
	ListRoadmaps(ctx context.Context, userID int64, archived bool) ([]models.RoadmapItem, error)
	// CountRoadmaps counts a user's non-archived roadmaps — backs the
	// MaxRoadmapsPerUser cap check.
	CountRoadmaps(ctx context.Context, userID int64) (int, error)
	GetRoadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error)
	RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error
	SetGoal(ctx context.Context, userID, roadmapID int64, goal string) error
	ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error
	ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error
	RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error
	DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error

	// Cards.
	AddCards(ctx context.Context, roadmapID int64, texts []string) (int, error)
	ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error)
	// ToggleCardDone flips one card's is_done (setting/clearing done_at) and
	// returns the roadmap it belongs to, so the caller can re-render the
	// right screen without carrying the roadmap id through the callback.
	ToggleCardDone(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	DeleteCard(ctx context.Context, userID, cardID int64) error

	// Push scheduling — mirrors LearningRepository's shape (migrations/0009).
	UpsertPushInterval(ctx context.Context, userID int64, intervalMin int, nextPushAt time.Time) error
	GetPushSettings(ctx context.Context, userID int64) (intervalMin int, nextPushAt time.Time, enabled bool, err error)
	SetNextPush(ctx context.Context, userID int64, nextPushAt time.Time) error
	DisablePush(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error)

	// PickDigestCards returns pending cards for a digest push, at most
	// perRoadmapCap per roadmap and totalCap overall.
	PickDigestCards(ctx context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error)

	// Stats.
	CountCards(ctx context.Context, userID int64) (total, done int, err error)
	GetRoadmapCardStats(ctx context.Context, userID int64) ([]models.RoadmapCardStat, error)
}

type roadmapRepository struct {
	db *pgxpool.Pool
}

// NewRoadmapRepository creates repository backed by pgx pool.
func NewRoadmapRepository(db *pgxpool.Pool) RoadmapRepository {
	return &roadmapRepository{db: db}
}

// CreateRoadmap inserts a new, active, non-archived roadmap with an empty
// goal.
func (r *roadmapRepository) CreateRoadmap(ctx context.Context, userID int64, name string) (int64, error) {
	q := `
	INSERT INTO roadmaps (user_id, name)
	VALUES ($1, $2)
	RETURNING id;
	`
	var id int64
	err := r.db.QueryRow(ctx, q, userID, name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, models.ErrRoadmapExists
		}
		return 0, fmt.Errorf("create roadmap: %w", err)
	}
	return id, nil
}

// ListRoadmaps returns a user's roadmaps (archived or not) with total/done
// card counts, ordered by creation time.
func (r *roadmapRepository) ListRoadmaps(ctx context.Context, userID int64, archived bool) ([]models.RoadmapItem, error) {
	q := `
	SELECT r.id, r.name, r.goal, r.is_active, r.is_archived,
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmaps r
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
	WHERE r.user_id = $1 AND r.is_archived = $2
	GROUP BY r.id
	ORDER BY r.created_at, r.id;
	`
	rows, err := r.db.Query(ctx, q, userID, archived)
	if err != nil {
		return nil, fmt.Errorf("list roadmaps query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapItem, 0)
	for rows.Next() {
		var item models.RoadmapItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Goal, &item.Active, &item.IsArchived, &item.TotalCards, &item.DoneCards); err != nil {
			return nil, fmt.Errorf("list roadmaps scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmaps rows: %w", err)
	}
	return out, nil
}

// CountRoadmaps counts a user's non-archived roadmaps.
func (r *roadmapRepository) CountRoadmaps(ctx context.Context, userID int64) (int, error) {
	q := `SELECT COUNT(*) FROM roadmaps WHERE user_id = $1 AND is_archived = FALSE;`
	var n int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count roadmaps: %w", err)
	}
	return n, nil
}

// GetRoadmap loads one roadmap with its card counts, scoped to its owner.
func (r *roadmapRepository) GetRoadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	q := `
	SELECT r.id, r.name, r.goal, r.is_active, r.is_archived,
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmaps r
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
	WHERE r.id = $1 AND r.user_id = $2
	GROUP BY r.id;
	`
	var item models.RoadmapItem
	err := r.db.QueryRow(ctx, q, roadmapID, userID).
		Scan(&item.ID, &item.Name, &item.Goal, &item.Active, &item.IsArchived, &item.TotalCards, &item.DoneCards)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RoadmapItem{}, models.ErrRoadmapNotFound
		}
		return models.RoadmapItem{}, fmt.Errorf("get roadmap: %w", err)
	}
	return item, nil
}

// RenameRoadmap updates a roadmap's display name, scoped to its owner.
func (r *roadmapRepository) RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error {
	q := `UPDATE roadmaps SET name = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID, newName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrRoadmapExists
		}
		return fmt.Errorf("rename roadmap: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// SetGoal replaces a roadmap's free-text mastery goal.
func (r *roadmapRepository) SetGoal(ctx context.Context, userID, roadmapID int64, goal string) error {
	q := `UPDATE roadmaps SET goal = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID, goal)
	if err != nil {
		return fmt.Errorf("set roadmap goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// ToggleRoadmapActive flips is_active (digest participation), scoped to the
// owning user.
func (r *roadmapRepository) ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error {
	q := `
	UPDATE roadmaps
	SET is_active = NOT is_active
	WHERE id = $1 AND user_id = $2 AND is_archived = FALSE;
	`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID)
	if err != nil {
		return fmt.Errorf("toggle roadmap active: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// ArchiveRoadmap moves a roadmap to the archive, freeing a slot in the
// MaxRoadmapsPerUser cap.
func (r *roadmapRepository) ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error {
	q := `UPDATE roadmaps SET is_archived = TRUE WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID)
	if err != nil {
		return fmt.Errorf("archive roadmap: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// RestoreRoadmap moves an archived roadmap back to the active list.
func (r *roadmapRepository) RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error {
	q := `UPDATE roadmaps SET is_archived = FALSE WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID)
	if err != nil {
		return fmt.Errorf("restore roadmap: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// DeleteRoadmapForever permanently removes a roadmap and its cards
// (ON DELETE CASCADE on roadmap_cards.roadmap_id).
func (r *roadmapRepository) DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error {
	q := `DELETE FROM roadmaps WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID)
	if err != nil {
		return fmt.Errorf("delete roadmap forever: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// AddCards bulk-inserts pending cards into a roadmap. Returns the number
// inserted.
func (r *roadmapRepository) AddCards(ctx context.Context, roadmapID int64, texts []string) (int, error) {
	if len(texts) == 0 {
		return 0, nil
	}
	batch := &pgx.Batch{}
	for _, text := range texts {
		batch.Queue(`INSERT INTO roadmap_cards (roadmap_id, text) VALUES ($1, $2);`, roadmapID, text)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range texts {
		if _, err := br.Exec(); err != nil {
			return 0, fmt.Errorf("add roadmap cards: %w", err)
		}
	}
	return len(texts), nil
}

// ListCards returns every card in one roadmap, scoped to its owner —
// pending first, then done, each group in insertion order, so the checklist
// reads as "what's left" without the finished items in the way.
//
// The id tiebreaker is load-bearing, not cosmetic: AddCards inserts a whole
// pasted block in one pgx batch, which runs as a single transaction, so
// every card in it shares the exact same created_at. Ordering by created_at
// alone would shuffle a pasted roadmap into an arbitrary order — and the
// order the user typed their plan in is the order they mean.
func (r *roadmapRepository) ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	q := `
	SELECT c.id, c.text, c.is_done, c.done_at
	FROM roadmap_cards c
	JOIN roadmaps r ON r.id = c.roadmap_id
	WHERE c.roadmap_id = $1 AND r.user_id = $2
	ORDER BY c.is_done, c.created_at, c.id;
	`
	rows, err := r.db.Query(ctx, q, roadmapID, userID)
	if err != nil {
		return nil, fmt.Errorf("list roadmap cards query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapCardItem, 0)
	for rows.Next() {
		var item models.RoadmapCardItem
		if err := rows.Scan(&item.ID, &item.Text, &item.IsDone, &item.DoneAt); err != nil {
			return nil, fmt.Errorf("list roadmap cards scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmap cards rows: %w", err)
	}
	return out, nil
}

// ToggleCardDone flips one card's done state, scoped to its owner via the
// roadmap join, and returns its roadmap id. done_at is set on the way to
// done and cleared on the way back to pending, so it always reflects the
// current state rather than "the last time it was ever ticked".
func (r *roadmapRepository) ToggleCardDone(ctx context.Context, userID, cardID int64) (int64, error) {
	q := `
	UPDATE roadmap_cards c
	SET is_done = NOT c.is_done,
		done_at = CASE WHEN c.is_done THEN NULL ELSE now() END
	FROM roadmaps r
	WHERE c.id = $1 AND r.id = c.roadmap_id AND r.user_id = $2
	RETURNING c.roadmap_id;
	`
	var roadmapID int64
	err := r.db.QueryRow(ctx, q, cardID, userID).Scan(&roadmapID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrRoadmapCardNotFound
		}
		return 0, fmt.Errorf("toggle roadmap card done: %w", err)
	}
	return roadmapID, nil
}

// DeleteCard removes one card, scoped to its owner via the roadmap join.
func (r *roadmapRepository) DeleteCard(ctx context.Context, userID, cardID int64) error {
	q := `
	DELETE FROM roadmap_cards c
	USING roadmaps r
	WHERE c.id = $1 AND c.roadmap_id = r.id AND r.user_id = $2;
	`
	tag, err := r.db.Exec(ctx, q, cardID, userID)
	if err != nil {
		return fmt.Errorf("delete roadmap card: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapCardNotFound
	}
	return nil
}

// UpsertPushInterval enables digest pushes and saves interval + next push
// timestamp — mirrors LearningRepository.UpsertPushInterval.
func (r *roadmapRepository) UpsertPushInterval(ctx context.Context, userID int64, intervalMin int, nextPushAt time.Time) error {
	q := `
	INSERT INTO user_roadmap_settings (user_id, interval_min, next_push_at, enabled, updated_at)
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
			return models.ErrRoadmapInvalidInterval
		}
		return fmt.Errorf("upsert roadmap push interval: %w", err)
	}
	return nil
}

// GetPushSettings returns the raw persisted row for user, if any. enabled is
// false and err is nil when the user has no row yet.
func (r *roadmapRepository) GetPushSettings(ctx context.Context, userID int64) (int, time.Time, bool, error) {
	q := `SELECT interval_min, next_push_at, enabled FROM user_roadmap_settings WHERE user_id = $1;`
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
		return 0, time.Time{}, false, fmt.Errorf("get roadmap push settings: %w", err)
	}
	return intervalMin, nextPushAt, enabled, nil
}

// SetNextPush updates next scheduled push time for user.
func (r *roadmapRepository) SetNextPush(ctx context.Context, userID int64, nextPushAt time.Time) error {
	q := `UPDATE user_roadmap_settings SET next_push_at = $2, updated_at = now() WHERE user_id = $1;`
	tag, err := r.db.Exec(ctx, q, userID, nextPushAt)
	if err != nil {
		return fmt.Errorf("set roadmap next push: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DisablePush turns off digest pushes for user.
func (r *roadmapRepository) DisablePush(ctx context.Context, userID int64) error {
	q := `UPDATE user_roadmap_settings SET enabled = FALSE, updated_at = now() WHERE user_id = $1;`
	_, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("disable roadmap push: %w", err)
	}
	return nil
}

// ListDueUsers returns users whose next_push_at is due.
func (r *roadmapRepository) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error) {
	q := `
	SELECT urs.user_id, u.tg_user_id, urs.interval_min
	FROM user_roadmap_settings urs
	JOIN users u ON u.id = urs.user_id
	WHERE urs.enabled = TRUE
	  AND urs.next_push_at IS NOT NULL
	  AND urs.next_push_at <= $1
	ORDER BY urs.next_push_at
	LIMIT $2;
	`
	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list roadmap due users query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapDueUser, 0, limit)
	for rows.Next() {
		var item models.RoadmapDueUser
		if err := rows.Scan(&item.DBUserID, &item.TgUserID, &item.IntervalMin); err != nil {
			return nil, fmt.Errorf("list roadmap due users scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmap due users rows: %w", err)
	}
	return out, nil
}

// PickDigestCards returns the oldest pending cards across the user's active,
// non-archived roadmaps — at most perRoadmapCap from any one roadmap (so a
// single long checklist can't monopolize the digest) and totalCap overall.
//
// "Oldest" is (created_at, id): a pasted batch shares one created_at (see
// ListCards), so id is what keeps both the per-roadmap pick and the overall
// truncation deterministic instead of returning a different arbitrary slice
// of the same block on every push.
func (r *roadmapRepository) PickDigestCards(ctx context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error) {
	q := `
	WITH pending AS (
		SELECT c.id, c.roadmap_id, r.name AS roadmap_name, c.text, c.created_at,
			ROW_NUMBER() OVER (PARTITION BY c.roadmap_id ORDER BY c.created_at, c.id) AS rn
		FROM roadmap_cards c
		JOIN roadmaps r ON r.id = c.roadmap_id
		WHERE r.user_id = $1 AND r.is_active = TRUE AND r.is_archived = FALSE
		  AND c.is_done = FALSE
	)
	SELECT id, roadmap_id, roadmap_name, text
	FROM pending
	WHERE rn <= $2
	ORDER BY created_at, id
	LIMIT $3;
	`
	rows, err := r.db.Query(ctx, q, userID, perRoadmapCap, totalCap)
	if err != nil {
		return nil, fmt.Errorf("pick digest cards query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapDigestCard, 0, totalCap)
	for rows.Next() {
		var item models.RoadmapDigestCard
		if err := rows.Scan(&item.ID, &item.RoadmapID, &item.RoadmapName, &item.Text); err != nil {
			return nil, fmt.Errorf("pick digest cards scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pick digest cards rows: %w", err)
	}
	return out, nil
}

// CountCards returns total/done card counts across all of a user's
// non-archived roadmaps.
func (r *roadmapRepository) CountCards(ctx context.Context, userID int64) (int, int, error) {
	q := `
	SELECT COUNT(*), COUNT(*) FILTER (WHERE c.is_done)
	FROM roadmap_cards c
	JOIN roadmaps r ON r.id = c.roadmap_id
	WHERE r.user_id = $1 AND r.is_archived = FALSE;
	`
	var total, done int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&total, &done); err != nil {
		return 0, 0, fmt.Errorf("count roadmap cards: %w", err)
	}
	return total, done, nil
}

// GetRoadmapCardStats returns per-roadmap card counts for the detailed
// "📈 Statistics" screen.
func (r *roadmapRepository) GetRoadmapCardStats(ctx context.Context, userID int64) ([]models.RoadmapCardStat, error) {
	q := `
	SELECT r.name, r.goal,
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmaps r
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
	WHERE r.user_id = $1 AND r.is_archived = FALSE
	GROUP BY r.id
	ORDER BY r.created_at, r.id;
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get roadmap card stats query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapCardStat, 0)
	for rows.Next() {
		var s models.RoadmapCardStat
		if err := rows.Scan(&s.Name, &s.Goal, &s.TotalCards, &s.DoneCards); err != nil {
			return nil, fmt.Errorf("get roadmap card stats scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get roadmap card stats rows: %w", err)
	}
	return out, nil
}
