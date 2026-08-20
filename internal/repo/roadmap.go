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

type RoadmapRepository interface {
	// Goals — the outcome a set of technologies feeds into.
	CreateGoal(ctx context.Context, userID int64, name string) (int64, error)
	ListGoals(ctx context.Context, userID int64, archived bool) ([]models.RoadmapGoalItem, error)
	CountGoals(ctx context.Context, userID int64) (int, error)
	GetGoal(ctx context.Context, userID, goalID int64) (models.RoadmapGoalItem, error)
	RenameGoal(ctx context.Context, userID, goalID int64, newName string) error
	ArchiveGoal(ctx context.Context, userID, goalID int64) error
	RestoreGoal(ctx context.Context, userID, goalID int64) error
	DeleteGoalForever(ctx context.Context, userID, goalID int64) error

	// Technologies.
	CreateRoadmap(ctx context.Context, userID, goalID int64, name string) (int64, error)
	// goalID nil lists technologies attached to no goal at all.
	ListRoadmaps(ctx context.Context, userID int64, goalID *int64, archived bool) ([]models.RoadmapItem, error)
	// Ignores which goal a technology belongs to — for the archive screen,
	// where grouping by goal would only get in the way.
	ListRoadmapsAnyGoal(ctx context.Context, userID int64, archived bool) ([]models.RoadmapItem, error)
	CountRoadmapsInGoal(ctx context.Context, userID, goalID int64) (int, error)
	CountRoadmaps(ctx context.Context, userID int64) (int, error)
	GetRoadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error)
	RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error
	SetMasteryCriteria(ctx context.Context, userID, roadmapID int64, criteria string) error
	AssignRoadmapToGoal(ctx context.Context, userID, roadmapID, goalID int64) error
	ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error
	ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error
	RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error
	DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error

	// Cards.
	AddCards(ctx context.Context, roadmapID int64, cards []models.RoadmapCardItem) (int, error)
	ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error)
	// Both return the owning roadmap, so the caller can re-render the right
	// screen without carrying it through the callback payload.
	ToggleCardDone(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	CycleCardDifficulty(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	DeleteCard(ctx context.Context, userID, cardID int64) error

	UpsertPushInterval(ctx context.Context, userID int64, intervalMin int, nextPushAt time.Time) error
	GetPushSettings(ctx context.Context, userID int64) (intervalMin int, nextPushAt time.Time, enabled bool, err error)
	SetNextPush(ctx context.Context, userID int64, nextPushAt time.Time) error
	DisablePush(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error)

	PickDigestCards(ctx context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error)

	CountCards(ctx context.Context, userID int64) (total, done int, err error)
	GetRoadmapCardStats(ctx context.Context, userID int64) ([]models.RoadmapCardStat, error)
}

type roadmapRepository struct {
	db *pgxpool.Pool
}

func NewRoadmapRepository(db *pgxpool.Pool) RoadmapRepository {
	return &roadmapRepository{db: db}
}

// The technology count is DISTINCT-ed because the double LEFT JOIN
// multiplies rows: one per (technology, card) pair.
const goalSelect = `
	SELECT g.id, g.name, g.is_archived,
		COUNT(DISTINCT r.id),
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmap_goals g
	LEFT JOIN roadmaps r ON r.goal_id = g.id AND r.is_archived = FALSE
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
`

func (r *roadmapRepository) CreateGoal(ctx context.Context, userID int64, name string) (int64, error) {
	q := `INSERT INTO roadmap_goals (user_id, name) VALUES ($1, $2) RETURNING id;`
	var id int64
	if err := r.db.QueryRow(ctx, q, userID, name).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, models.ErrRoadmapGoalExists
		}
		return 0, fmt.Errorf("create goal: %w", err)
	}
	return id, nil
}

func (r *roadmapRepository) ListGoals(ctx context.Context, userID int64, archived bool) ([]models.RoadmapGoalItem, error) {
	q := goalSelect + `
	WHERE g.user_id = $1 AND g.is_archived = $2
	GROUP BY g.id
	ORDER BY g.created_at, g.id;
	`
	rows, err := r.db.Query(ctx, q, userID, archived)
	if err != nil {
		return nil, fmt.Errorf("list goals query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapGoalItem, 0)
	for rows.Next() {
		var item models.RoadmapGoalItem
		if err := rows.Scan(&item.ID, &item.Name, &item.IsArchived, &item.TotalRoadmaps, &item.TotalCards, &item.DoneCards); err != nil {
			return nil, fmt.Errorf("list goals scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list goals rows: %w", err)
	}
	return out, nil
}

func (r *roadmapRepository) CountGoals(ctx context.Context, userID int64) (int, error) {
	q := `SELECT COUNT(*) FROM roadmap_goals WHERE user_id = $1 AND is_archived = FALSE;`
	var n int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count goals: %w", err)
	}
	return n, nil
}

func (r *roadmapRepository) GetGoal(ctx context.Context, userID, goalID int64) (models.RoadmapGoalItem, error) {
	q := goalSelect + `
	WHERE g.id = $1 AND g.user_id = $2
	GROUP BY g.id;
	`
	var item models.RoadmapGoalItem
	err := r.db.QueryRow(ctx, q, goalID, userID).
		Scan(&item.ID, &item.Name, &item.IsArchived, &item.TotalRoadmaps, &item.TotalCards, &item.DoneCards)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RoadmapGoalItem{}, models.ErrRoadmapGoalNotFound
		}
		return models.RoadmapGoalItem{}, fmt.Errorf("get goal: %w", err)
	}
	return item, nil
}

func (r *roadmapRepository) RenameGoal(ctx context.Context, userID, goalID int64, newName string) error {
	q := `UPDATE roadmap_goals SET name = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, goalID, userID, newName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.ErrRoadmapGoalExists
		}
		return fmt.Errorf("rename goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapGoalNotFound
	}
	return nil
}

func (r *roadmapRepository) ArchiveGoal(ctx context.Context, userID, goalID int64) error {
	return r.setGoalArchived(ctx, userID, goalID, true)
}

func (r *roadmapRepository) RestoreGoal(ctx context.Context, userID, goalID int64) error {
	return r.setGoalArchived(ctx, userID, goalID, false)
}

func (r *roadmapRepository) setGoalArchived(ctx context.Context, userID, goalID int64, archived bool) error {
	q := `UPDATE roadmap_goals SET is_archived = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, goalID, userID, archived)
	if err != nil {
		return fmt.Errorf("set goal archived: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapGoalNotFound
	}
	return nil
}

// The FK is ON DELETE SET NULL, so the goal's technologies survive as
// unattached instead of disappearing with it.
func (r *roadmapRepository) DeleteGoalForever(ctx context.Context, userID, goalID int64) error {
	q := `DELETE FROM roadmap_goals WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, goalID, userID)
	if err != nil {
		return fmt.Errorf("delete goal forever: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapGoalNotFound
	}
	return nil
}

const roadmapSelect = `
	SELECT r.id, r.goal_id, r.name, r.mastery_criteria, r.is_active, r.is_archived,
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmaps r
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
`

func (r *roadmapRepository) CreateRoadmap(ctx context.Context, userID, goalID int64, name string) (int64, error) {
	q := `
	INSERT INTO roadmaps (user_id, goal_id, name)
	SELECT $1, $2, $3
	WHERE EXISTS (SELECT 1 FROM roadmap_goals WHERE id = $2 AND user_id = $1)
	RETURNING id;
	`
	var id int64
	err := r.db.QueryRow(ctx, q, userID, goalID, name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, models.ErrRoadmapExists
		}
		// No row inserted means WHERE EXISTS failed: the goal isn't this
		// user's, so it's an ownership miss rather than a DB error.
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrRoadmapGoalNotFound
		}
		return 0, fmt.Errorf("create roadmap: %w", err)
	}
	return id, nil
}

func (r *roadmapRepository) ListRoadmaps(ctx context.Context, userID int64, goalID *int64, archived bool) ([]models.RoadmapItem, error) {
	q := roadmapSelect + `
	WHERE r.user_id = $1 AND r.is_archived = $2
	  AND (($3::BIGINT IS NULL AND r.goal_id IS NULL) OR r.goal_id = $3)
	GROUP BY r.id
	ORDER BY r.created_at, r.id;
	`
	rows, err := r.db.Query(ctx, q, userID, archived, goalID)
	if err != nil {
		return nil, fmt.Errorf("list roadmaps query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapItem, 0)
	for rows.Next() {
		var item models.RoadmapItem
		if err := rows.Scan(&item.ID, &item.GoalID, &item.Name, &item.MasteryCriteria, &item.Active, &item.IsArchived, &item.TotalCards, &item.DoneCards); err != nil {
			return nil, fmt.Errorf("list roadmaps scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmaps rows: %w", err)
	}
	return out, nil
}

func (r *roadmapRepository) ListRoadmapsAnyGoal(ctx context.Context, userID int64, archived bool) ([]models.RoadmapItem, error) {
	q := roadmapSelect + `
	WHERE r.user_id = $1 AND r.is_archived = $2
	GROUP BY r.id
	ORDER BY r.created_at, r.id;
	`
	rows, err := r.db.Query(ctx, q, userID, archived)
	if err != nil {
		return nil, fmt.Errorf("list roadmaps any goal query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapItem, 0)
	for rows.Next() {
		var item models.RoadmapItem
		if err := rows.Scan(&item.ID, &item.GoalID, &item.Name, &item.MasteryCriteria, &item.Active, &item.IsArchived, &item.TotalCards, &item.DoneCards); err != nil {
			return nil, fmt.Errorf("list roadmaps any goal scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmaps any goal rows: %w", err)
	}
	return out, nil
}

func (r *roadmapRepository) CountRoadmapsInGoal(ctx context.Context, userID, goalID int64) (int, error) {
	q := `SELECT COUNT(*) FROM roadmaps WHERE user_id = $1 AND goal_id = $2 AND is_archived = FALSE;`
	var n int
	if err := r.db.QueryRow(ctx, q, userID, goalID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count roadmaps in goal: %w", err)
	}
	return n, nil
}

func (r *roadmapRepository) CountRoadmaps(ctx context.Context, userID int64) (int, error) {
	q := `SELECT COUNT(*) FROM roadmaps WHERE user_id = $1 AND is_archived = FALSE;`
	var n int
	if err := r.db.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count roadmaps: %w", err)
	}
	return n, nil
}

func (r *roadmapRepository) GetRoadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	q := roadmapSelect + `
	WHERE r.id = $1 AND r.user_id = $2
	GROUP BY r.id;
	`
	var item models.RoadmapItem
	err := r.db.QueryRow(ctx, q, roadmapID, userID).
		Scan(&item.ID, &item.GoalID, &item.Name, &item.MasteryCriteria, &item.Active, &item.IsArchived, &item.TotalCards, &item.DoneCards)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RoadmapItem{}, models.ErrRoadmapNotFound
		}
		return models.RoadmapItem{}, fmt.Errorf("get roadmap: %w", err)
	}
	return item, nil
}

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

func (r *roadmapRepository) SetMasteryCriteria(ctx context.Context, userID, roadmapID int64, criteria string) error {
	q := `UPDATE roadmaps SET mastery_criteria = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID, criteria)
	if err != nil {
		return fmt.Errorf("set mastery criteria: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

// Scoped both ways, since the goal id arrives from a callback payload: a
// user can only move their own technology into their own goal.
func (r *roadmapRepository) AssignRoadmapToGoal(ctx context.Context, userID, roadmapID, goalID int64) error {
	q := `
	UPDATE roadmaps SET goal_id = $3
	WHERE id = $1 AND user_id = $2
	  AND EXISTS (SELECT 1 FROM roadmap_goals WHERE id = $3 AND user_id = $2);
	`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID, goalID)
	if err != nil {
		return fmt.Errorf("assign roadmap to goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

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

func (r *roadmapRepository) ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error {
	return r.setRoadmapArchived(ctx, userID, roadmapID, true)
}

func (r *roadmapRepository) RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error {
	return r.setRoadmapArchived(ctx, userID, roadmapID, false)
}

func (r *roadmapRepository) setRoadmapArchived(ctx context.Context, userID, roadmapID int64, archived bool) error {
	q := `UPDATE roadmaps SET is_archived = $3 WHERE id = $1 AND user_id = $2;`
	tag, err := r.db.Exec(ctx, q, roadmapID, userID, archived)
	if err != nil {
		return fmt.Errorf("set roadmap archived: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrRoadmapNotFound
	}
	return nil
}

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

func (r *roadmapRepository) AddCards(ctx context.Context, roadmapID int64, cards []models.RoadmapCardItem) (int, error) {
	if len(cards) == 0 {
		return 0, nil
	}
	batch := &pgx.Batch{}
	for _, c := range cards {
		batch.Queue(`INSERT INTO roadmap_cards (roadmap_id, text, kind, difficulty) VALUES ($1, $2, $3, $4);`,
			roadmapID, c.Text, string(c.Kind), c.Difficulty)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range cards {
		if _, err := br.Exec(); err != nil {
			return 0, fmt.Errorf("add roadmap cards: %w", err)
		}
	}
	return len(cards), nil
}

// Pending first, easiest-first within that, so the checklist reads as "what
// to do next" rather than "what I typed first". The id tiebreaker is
// load-bearing: a pasted block inserts as one batch sharing a single
// created_at, so without it the order inside a difficulty tier is arbitrary.
func (r *roadmapRepository) ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	q := `
	SELECT c.id, c.text, c.kind, c.difficulty, c.is_done, c.done_at
	FROM roadmap_cards c
	JOIN roadmaps r ON r.id = c.roadmap_id
	WHERE c.roadmap_id = $1 AND r.user_id = $2
	ORDER BY c.is_done, c.difficulty, c.created_at, c.id;
	`
	rows, err := r.db.Query(ctx, q, roadmapID, userID)
	if err != nil {
		return nil, fmt.Errorf("list roadmap cards query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapCardItem, 0)
	for rows.Next() {
		var item models.RoadmapCardItem
		if err := rows.Scan(&item.ID, &item.Text, &item.Kind, &item.Difficulty, &item.IsDone, &item.DoneAt); err != nil {
			return nil, fmt.Errorf("list roadmap cards scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list roadmap cards rows: %w", err)
	}
	return out, nil
}

// done_at is cleared on the way back to pending, so it always reflects the
// current state rather than "the last time this was ever ticked".
func (r *roadmapRepository) ToggleCardDone(ctx context.Context, userID, cardID int64) (int64, error) {
	q := `
	UPDATE roadmap_cards c
	SET is_done = NOT c.is_done,
		done_at = CASE WHEN c.is_done THEN NULL ELSE now() END
	FROM roadmaps r
	WHERE c.id = $1 AND r.id = c.roadmap_id AND r.user_id = $2
	RETURNING c.roadmap_id;
	`
	return r.cardMutation(ctx, q, cardID, userID)
}

// Cycles 1 -> 2 -> 3 -> 1. A cycle rather than a picker screen: there are
// only three values, and re-tapping one button beats navigating into a card
// and back out.
func (r *roadmapRepository) CycleCardDifficulty(ctx context.Context, userID, cardID int64) (int64, error) {
	q := `
	UPDATE roadmap_cards c
	SET difficulty = (c.difficulty % 3) + 1
	FROM roadmaps r
	WHERE c.id = $1 AND r.id = c.roadmap_id AND r.user_id = $2
	RETURNING c.roadmap_id;
	`
	return r.cardMutation(ctx, q, cardID, userID)
}

func (r *roadmapRepository) cardMutation(ctx context.Context, q string, cardID, userID int64) (int64, error) {
	var roadmapID int64
	if err := r.db.QueryRow(ctx, q, cardID, userID).Scan(&roadmapID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, models.ErrRoadmapCardNotFound
		}
		return 0, fmt.Errorf("roadmap card mutation: %w", err)
	}
	return roadmapID, nil
}

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
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return models.ErrRoadmapInvalidInterval
		}
		return fmt.Errorf("upsert roadmap push interval: %w", err)
	}
	return nil
}

// enabled is false with a nil error when the user has no row yet.
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

func (r *roadmapRepository) DisablePush(ctx context.Context, userID int64) error {
	q := `UPDATE user_roadmap_settings SET enabled = FALSE, updated_at = now() WHERE user_id = $1;`
	if _, err := r.db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("disable roadmap push: %w", err)
	}
	return nil
}

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

// Easiest pending cards first — the digest's job is to offer the next
// realistic step, not the oldest one. Capped per technology so one long
// checklist can't monopolize a push; an archived goal takes its
// technologies out of the rotation with it.
func (r *roadmapRepository) PickDigestCards(ctx context.Context, userID int64, perRoadmapCap, totalCap int) ([]models.RoadmapDigestCard, error) {
	q := `
	WITH pending AS (
		SELECT c.id, c.roadmap_id, r.name AS roadmap_name, c.text, c.kind, c.difficulty, c.created_at,
			ROW_NUMBER() OVER (PARTITION BY c.roadmap_id ORDER BY c.difficulty, c.created_at, c.id) AS rn
		FROM roadmap_cards c
		JOIN roadmaps r ON r.id = c.roadmap_id
		LEFT JOIN roadmap_goals g ON g.id = r.goal_id
		WHERE r.user_id = $1 AND r.is_active = TRUE AND r.is_archived = FALSE
		  AND c.is_done = FALSE
		  AND (g.id IS NULL OR g.is_archived = FALSE)
	)
	SELECT id, roadmap_id, roadmap_name, text, kind, difficulty
	FROM pending
	WHERE rn <= $2
	ORDER BY difficulty, created_at, id
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
		if err := rows.Scan(&item.ID, &item.RoadmapID, &item.RoadmapName, &item.Text, &item.Kind, &item.Difficulty); err != nil {
			return nil, fmt.Errorf("pick digest cards scan: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pick digest cards rows: %w", err)
	}
	return out, nil
}

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

// GoalName is empty for a technology attached to no goal.
func (r *roadmapRepository) GetRoadmapCardStats(ctx context.Context, userID int64) ([]models.RoadmapCardStat, error) {
	q := `
	SELECT COALESCE(g.name, ''), r.name, r.mastery_criteria,
		COUNT(c.id),
		COUNT(c.id) FILTER (WHERE c.is_done)
	FROM roadmaps r
	LEFT JOIN roadmap_goals g ON g.id = r.goal_id
	LEFT JOIN roadmap_cards c ON c.roadmap_id = r.id
	WHERE r.user_id = $1 AND r.is_archived = FALSE
	GROUP BY r.id, g.name, g.created_at
	ORDER BY g.created_at NULLS LAST, r.created_at, r.id;
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get roadmap card stats query: %w", err)
	}
	defer rows.Close()

	out := make([]models.RoadmapCardStat, 0)
	for rows.Next() {
		var s models.RoadmapCardStat
		if err := rows.Scan(&s.GoalName, &s.Name, &s.MasteryCriteria, &s.TotalCards, &s.DoneCards); err != nil {
			return nil, fmt.Errorf("get roadmap card stats scan: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get roadmap card stats rows: %w", err)
	}
	return out, nil
}
