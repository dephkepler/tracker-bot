package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// RoadmapService contains learning-roadmap use-cases: up to
// models.MaxRoadmapsPerUser technologies, each with a free-text mastery goal
// and a checklist of freeform cards, nudged by a periodic digest push
// (mirrors LearningService's Activate/Stop/ListDueUsers/MarkPushSent shape).
type RoadmapService interface {
	CreateRoadmap(ctx context.Context, userID int64, name string) (int64, error)
	ListRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error)
	ListArchivedRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error)
	Roadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error)
	RoadmapName(ctx context.Context, userID, roadmapID int64) (string, error)
	RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error
	// SetGoal replaces a roadmap's free-text mastery goal ("what does
	// 'I know this' mean for me"). An empty goal is valid — the create flow
	// lets the user skip it.
	SetGoal(ctx context.Context, userID, roadmapID int64, goal string) error
	ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error
	ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error
	RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error
	DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error

	// AddCardsFromText parses free-text lines (one card per line) and
	// appends them to a roadmap. Returns how many lines became cards versus
	// were skipped (blank after trimming, or longer than
	// models.MaxRoadmapCardTextLen).
	AddCardsFromText(ctx context.Context, userID, roadmapID int64, text string) (added, skipped int, err error)
	ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error)
	// ToggleCardDone flips one card's done state and returns its roadmap id,
	// so the caller can re-render the right screen — a card can be ticked
	// from its roadmap's checklist or straight off a digest push.
	ToggleCardDone(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	DeleteCard(ctx context.Context, userID, cardID int64) error

	// Push scheduling.
	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error)
	MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error

	// PickDigestCards returns the pending cards for one digest push, capped
	// per roadmap and overall (see models.RoadmapDigest* constants).
	PickDigestCards(ctx context.Context, userID int64) ([]models.RoadmapDigestCard, error)

	GetRoadmapStats(ctx context.Context, userID int64) (models.RoadmapStats, error)
	// GetStatsDetail returns the full breakdown behind the "📈 Statistics"
	// screen: overall numbers plus a per-roadmap table.
	GetStatsDetail(ctx context.Context, userID int64) (models.RoadmapStatsDetail, error)
}

type roadmapService struct {
	repo repo.RoadmapRepository
}

func NewRoadmapService(repo repo.RoadmapRepository) RoadmapService {
	return &roadmapService{
		repo: repo,
	}
}

// CreateRoadmap validates the name, enforces the MaxRoadmapsPerUser cap and
// stores a new roadmap.
//
// The cap is a count-then-insert without a transaction, so two concurrent
// creates could in principle both pass the check — the same unguarded
// pattern timerService.AddCustomInterval already uses for
// MaxCustomTimersPerUser, and equally harmless here: the worst case is one
// extra roadmap for a user racing themselves across two Telegram clients.
func (srv *roadmapService) CreateRoadmap(ctx context.Context, userID int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 {
		return 0, fmt.Errorf("create roadmap: invalid userID")
	}
	if strings.Contains(name, "\n") || len(name) < 2 || len(name) > 60 {
		return 0, models.ErrRoadmapInvalidName
	}

	count, err := srv.repo.CountRoadmaps(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("create roadmap: %w", err)
	}
	if count >= models.MaxRoadmapsPerUser {
		return 0, models.ErrRoadmapLimitReached
	}

	return srv.repo.CreateRoadmap(ctx, userID, name)
}

// ListRoadmaps returns a user's active (non-archived) roadmaps.
func (srv *roadmapService) ListRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error) {
	return srv.repo.ListRoadmaps(ctx, userID, false)
}

// ListArchivedRoadmaps returns a user's archived roadmaps.
func (srv *roadmapService) ListArchivedRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error) {
	return srv.repo.ListRoadmaps(ctx, userID, true)
}

// Roadmap loads one roadmap with its card counts.
func (srv *roadmapService) Roadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	return srv.repo.GetRoadmap(ctx, userID, roadmapID)
}

// RoadmapName resolves a roadmap's display name.
func (srv *roadmapService) RoadmapName(ctx context.Context, userID, roadmapID int64) (string, error) {
	item, err := srv.repo.GetRoadmap(ctx, userID, roadmapID)
	if err != nil {
		return "", err
	}
	return item.Name, nil
}

// RenameRoadmap validates and applies a new name for a roadmap.
func (srv *roadmapService) RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if strings.Contains(newName, "\n") || len(newName) < 2 || len(newName) > 60 {
		return models.ErrRoadmapInvalidName
	}
	return srv.repo.RenameRoadmap(ctx, userID, roadmapID, newName)
}

// SetGoal validates and stores a roadmap's mastery goal. Newlines are
// collapsed to spaces so the goal stays a single display line on every
// screen that shows it.
func (srv *roadmapService) SetGoal(ctx context.Context, userID, roadmapID int64, goal string) error {
	goal = strings.TrimSpace(strings.ReplaceAll(goal, "\n", " "))
	if len([]rune(goal)) > models.MaxRoadmapGoalLen {
		return models.ErrRoadmapGoalTooLong
	}
	return srv.repo.SetGoal(ctx, userID, roadmapID, goal)
}

// ToggleRoadmapActive flips whether a roadmap participates in digest pushes.
func (srv *roadmapService) ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.ToggleRoadmapActive(ctx, userID, roadmapID)
}

// ArchiveRoadmap moves a roadmap to the archive, freeing a slot in the cap.
func (srv *roadmapService) ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.ArchiveRoadmap(ctx, userID, roadmapID)
}

// RestoreRoadmap moves an archived roadmap back to the active list. It
// re-occupies a slot, so the cap is re-checked here too.
func (srv *roadmapService) RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error {
	count, err := srv.repo.CountRoadmaps(ctx, userID)
	if err != nil {
		return fmt.Errorf("restore roadmap: %w", err)
	}
	if count >= models.MaxRoadmapsPerUser {
		return models.ErrRoadmapLimitReached
	}
	return srv.repo.RestoreRoadmap(ctx, userID, roadmapID)
}

// DeleteRoadmapForever permanently removes a roadmap and its cards.
func (srv *roadmapService) DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.DeleteRoadmapForever(ctx, userID, roadmapID)
}

// AddCardsFromText turns one pasted block into cards, one per non-blank
// line. Unlike Learning's word lists there's no separator to parse — a card
// is whatever the user typed — so the only way a line is skipped is being
// blank or over the length cap.
func (srv *roadmapService) AddCardsFromText(ctx context.Context, userID, roadmapID int64, text string) (int, int, error) {
	if _, err := srv.repo.GetRoadmap(ctx, userID, roadmapID); err != nil {
		return 0, 0, err
	}

	texts := make([]string, 0)
	skipped := 0
	for _, line := range strings.Split(text, "\n") {
		card := normalizeCardLine(line)
		if card == "" {
			continue
		}
		if len([]rune(card)) > models.MaxRoadmapCardTextLen {
			// Skipped rather than truncated: silently cutting a line in
			// half loses the user's text with no sign it happened, while a
			// skip is reported back as "(N line(s) skipped)".
			skipped++
			continue
		}
		texts = append(texts, card)
	}

	if len(texts) == 0 {
		return 0, skipped, models.ErrRoadmapNoCardsParsed
	}

	added, err := srv.repo.AddCards(ctx, roadmapID, texts)
	if err != nil {
		return 0, skipped, err
	}
	return added, skipped, nil
}

// cardLineMarkers are leading list markers stripped from a pasted line —
// people paste checklists straight out of notes/docs, and keeping the "- "
// would double up with the bullet the UI draws itself.
var cardLineMarkers = []string{"- ", "— ", "– ", "* ", "• ", "· "}

// normalizeCardLine trims a line and strips one leading list marker.
// Returns "" for a line that has nothing left.
func normalizeCardLine(line string) string {
	line = strings.TrimSpace(line)
	for _, marker := range cardLineMarkers {
		if strings.HasPrefix(line, marker) {
			line = strings.TrimSpace(strings.TrimPrefix(line, marker))
			break
		}
	}
	return line
}

// ListCards returns every card in one roadmap, pending first.
func (srv *roadmapService) ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	return srv.repo.ListCards(ctx, userID, roadmapID)
}

// ToggleCardDone flips one card's done state.
func (srv *roadmapService) ToggleCardDone(ctx context.Context, userID, cardID int64) (int64, error) {
	return srv.repo.ToggleCardDone(ctx, userID, cardID)
}

// DeleteCard removes one card.
func (srv *roadmapService) DeleteCard(ctx context.Context, userID, cardID int64) error {
	return srv.repo.DeleteCard(ctx, userID, cardID)
}

// Activate enables digest pushes and schedules the next one. Mirrors
// LearningService.Activate: an already-running schedule with the same
// interval is kept as-is rather than restarted.
func (srv *roadmapService) Activate(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 {
		return fmt.Errorf("activate roadmap push: invalid userID")
	}
	if intervalMin <= 0 || intervalMin > 1440 {
		return models.ErrRoadmapInvalidInterval
	}

	now := time.Now().UTC()
	nextPushAt := now.Add(time.Duration(intervalMin) * time.Minute)

	curInterval, curNextPushAt, active, err := srv.repo.GetPushSettings(ctx, userID)
	if err != nil {
		return fmt.Errorf("activate roadmap push: %w", err)
	}
	if active && curInterval == intervalMin && curNextPushAt.After(now) {
		nextPushAt = curNextPushAt
	}

	return srv.repo.UpsertPushInterval(ctx, userID, intervalMin, nextPushAt)
}

// Stop disables digest pushes for user.
func (srv *roadmapService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop roadmap push: invalid userID")
	}
	return srv.repo.DisablePush(ctx, userID)
}

// ListDueUsers returns users that should receive a digest push now.
func (srv *roadmapService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return srv.repo.ListDueUsers(ctx, now.UTC(), limit)
}

// MarkPushSent moves next push time forward by interval.
func (srv *roadmapService) MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPushAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return srv.repo.SetNextPush(ctx, userID, nextPushAt)
}

// PickDigestCards returns the pending cards for one digest push.
func (srv *roadmapService) PickDigestCards(ctx context.Context, userID int64) ([]models.RoadmapDigestCard, error) {
	return srv.repo.PickDigestCards(ctx, userID, models.RoadmapDigestPerRoadmapCap, models.RoadmapDigestMaxCards)
}

// GetRoadmapStats aggregates the Roadmap dashboard's numbers.
func (srv *roadmapService) GetRoadmapStats(ctx context.Context, userID int64) (models.RoadmapStats, error) {
	items, err := srv.repo.ListRoadmaps(ctx, userID, false)
	if err != nil {
		return models.RoadmapStats{}, err
	}

	total, done, err := srv.repo.CountCards(ctx, userID)
	if err != nil {
		return models.RoadmapStats{}, err
	}

	stats := models.RoadmapStats{
		TotalRoadmaps: len(items),
		TotalCards:    total,
		DoneCards:     done,
		PendingCards:  total - done,
	}

	intervalMin, nextPushAt, active, err := srv.repo.GetPushSettings(ctx, userID)
	if err != nil {
		return models.RoadmapStats{}, err
	}
	stats.TimerActive = active
	stats.TimerInterval = intervalMin
	if active {
		remaining := time.Until(nextPushAt)
		if remaining < 0 {
			remaining = 0
		}
		stats.NextPushIn = fmt.Sprintf("%d min", int(remaining.Minutes()))
	}
	return stats, nil
}

// GetStatsDetail builds the full "📈 Statistics" breakdown: overall numbers
// plus a per-roadmap table.
func (srv *roadmapService) GetStatsDetail(ctx context.Context, userID int64) (models.RoadmapStatsDetail, error) {
	overall, err := srv.GetRoadmapStats(ctx, userID)
	if err != nil {
		return models.RoadmapStatsDetail{}, err
	}

	roadmaps, err := srv.repo.GetRoadmapCardStats(ctx, userID)
	if err != nil {
		return models.RoadmapStatsDetail{}, err
	}

	return models.RoadmapStatsDetail{
		Overall:  overall,
		Roadmaps: roadmaps,
	}, nil
}
