package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// RoadmapService models a three-level learning plan: an outcome the user is
// working toward ("reach mid-level"), the technologies that feed it, and a
// checklist of freeform cards under each technology. Cards carry a
// difficulty so the plan can be walked easiest-first, and a periodic digest
// push offers the next steps.
type RoadmapService interface {
	CreateGoal(ctx context.Context, userID int64, name string) (int64, error)
	ListGoals(ctx context.Context, userID int64) ([]models.RoadmapGoalItem, error)
	ListArchivedGoals(ctx context.Context, userID int64) ([]models.RoadmapGoalItem, error)
	Goal(ctx context.Context, userID, goalID int64) (models.RoadmapGoalItem, error)
	RenameGoal(ctx context.Context, userID, goalID int64, newName string) error
	ArchiveGoal(ctx context.Context, userID, goalID int64) error
	RestoreGoal(ctx context.Context, userID, goalID int64) error
	DeleteGoalForever(ctx context.Context, userID, goalID int64) error

	CreateRoadmap(ctx context.Context, userID, goalID int64, name string) (int64, error)
	ListRoadmaps(ctx context.Context, userID, goalID int64) ([]models.RoadmapItem, error)
	// Technologies with no goal — v1 leftovers, or ones whose goal was
	// deleted (the FK nulls them rather than deleting them).
	ListOrphanRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error)
	ListArchivedRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error)
	Roadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error)
	RoadmapName(ctx context.Context, userID, roadmapID int64) (string, error)
	RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error
	SetMasteryCriteria(ctx context.Context, userID, roadmapID int64, criteria string) error
	AssignRoadmapToGoal(ctx context.Context, userID, roadmapID, goalID int64) error
	ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error
	ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error
	RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error
	DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error

	// AddCardsFromText parses one card per non-blank line, honouring inline
	// #kind and !difficulty tags (see parseCardLine).
	AddCardsFromText(ctx context.Context, userID, roadmapID int64, text string) (added, skipped int, err error)
	ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error)
	ToggleCardDone(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	CycleCardDifficulty(ctx context.Context, userID, cardID int64) (roadmapID int64, err error)
	DeleteCard(ctx context.Context, userID, cardID int64) error

	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error)
	MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error

	PickDigestCards(ctx context.Context, userID int64) ([]models.RoadmapDigestCard, error)

	GetRoadmapStats(ctx context.Context, userID int64) (models.RoadmapStats, error)
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

// Both caps below are count-then-insert without a transaction, the same
// unguarded pattern timerService.AddCustomInterval already uses: worst case
// a user racing themself across two clients gets one extra row.
func (srv *roadmapService) CreateGoal(ctx context.Context, userID int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 {
		return 0, fmt.Errorf("create goal: invalid userID")
	}
	if !validRoadmapName(name) {
		return 0, models.ErrRoadmapInvalidName
	}

	count, err := srv.repo.CountGoals(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("create goal: %w", err)
	}
	if count >= models.MaxRoadmapGoalsPerUser {
		return 0, models.ErrRoadmapGoalLimitReached
	}
	return srv.repo.CreateGoal(ctx, userID, name)
}

func (srv *roadmapService) ListGoals(ctx context.Context, userID int64) ([]models.RoadmapGoalItem, error) {
	return srv.repo.ListGoals(ctx, userID, false)
}

func (srv *roadmapService) ListArchivedGoals(ctx context.Context, userID int64) ([]models.RoadmapGoalItem, error) {
	return srv.repo.ListGoals(ctx, userID, true)
}

func (srv *roadmapService) Goal(ctx context.Context, userID, goalID int64) (models.RoadmapGoalItem, error) {
	return srv.repo.GetGoal(ctx, userID, goalID)
}

func (srv *roadmapService) RenameGoal(ctx context.Context, userID, goalID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if !validRoadmapName(newName) {
		return models.ErrRoadmapInvalidName
	}
	return srv.repo.RenameGoal(ctx, userID, goalID, newName)
}

func (srv *roadmapService) ArchiveGoal(ctx context.Context, userID, goalID int64) error {
	return srv.repo.ArchiveGoal(ctx, userID, goalID)
}

// Restoring re-occupies a slot, so the cap is re-checked here too.
func (srv *roadmapService) RestoreGoal(ctx context.Context, userID, goalID int64) error {
	count, err := srv.repo.CountGoals(ctx, userID)
	if err != nil {
		return fmt.Errorf("restore goal: %w", err)
	}
	if count >= models.MaxRoadmapGoalsPerUser {
		return models.ErrRoadmapGoalLimitReached
	}
	return srv.repo.RestoreGoal(ctx, userID, goalID)
}

func (srv *roadmapService) DeleteGoalForever(ctx context.Context, userID, goalID int64) error {
	return srv.repo.DeleteGoalForever(ctx, userID, goalID)
}

func (srv *roadmapService) CreateRoadmap(ctx context.Context, userID, goalID int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 {
		return 0, fmt.Errorf("create roadmap: invalid userID")
	}
	if !validRoadmapName(name) {
		return 0, models.ErrRoadmapInvalidName
	}

	count, err := srv.repo.CountRoadmapsInGoal(ctx, userID, goalID)
	if err != nil {
		return 0, fmt.Errorf("create roadmap: %w", err)
	}
	if count >= models.MaxRoadmapsPerGoal {
		return 0, models.ErrRoadmapLimitReached
	}
	return srv.repo.CreateRoadmap(ctx, userID, goalID, name)
}

func (srv *roadmapService) ListRoadmaps(ctx context.Context, userID, goalID int64) ([]models.RoadmapItem, error) {
	return srv.repo.ListRoadmaps(ctx, userID, &goalID, false)
}

func (srv *roadmapService) ListOrphanRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error) {
	return srv.repo.ListRoadmaps(ctx, userID, nil, false)
}

func (srv *roadmapService) ListArchivedRoadmaps(ctx context.Context, userID int64) ([]models.RoadmapItem, error) {
	return srv.repo.ListRoadmapsAnyGoal(ctx, userID, true)
}

func (srv *roadmapService) Roadmap(ctx context.Context, userID, roadmapID int64) (models.RoadmapItem, error) {
	return srv.repo.GetRoadmap(ctx, userID, roadmapID)
}

func (srv *roadmapService) RoadmapName(ctx context.Context, userID, roadmapID int64) (string, error) {
	item, err := srv.repo.GetRoadmap(ctx, userID, roadmapID)
	if err != nil {
		return "", err
	}
	return item.Name, nil
}

func (srv *roadmapService) RenameRoadmap(ctx context.Context, userID, roadmapID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if !validRoadmapName(newName) {
		return models.ErrRoadmapInvalidName
	}
	return srv.repo.RenameRoadmap(ctx, userID, roadmapID, newName)
}

// Newlines collapse to spaces so the criteria stays one display line on
// every screen that shows it.
func (srv *roadmapService) SetMasteryCriteria(ctx context.Context, userID, roadmapID int64, criteria string) error {
	criteria = strings.TrimSpace(strings.ReplaceAll(criteria, "\n", " "))
	if len([]rune(criteria)) > models.MaxRoadmapCriteriaLen {
		return models.ErrRoadmapCriteriaTooLong
	}
	return srv.repo.SetMasteryCriteria(ctx, userID, roadmapID, criteria)
}

// Moving a technology into a goal fills a slot there, so the target goal's
// cap applies just like it does on create.
func (srv *roadmapService) AssignRoadmapToGoal(ctx context.Context, userID, roadmapID, goalID int64) error {
	count, err := srv.repo.CountRoadmapsInGoal(ctx, userID, goalID)
	if err != nil {
		return fmt.Errorf("assign roadmap to goal: %w", err)
	}
	if count >= models.MaxRoadmapsPerGoal {
		return models.ErrRoadmapLimitReached
	}
	return srv.repo.AssignRoadmapToGoal(ctx, userID, roadmapID, goalID)
}

func (srv *roadmapService) ToggleRoadmapActive(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.ToggleRoadmapActive(ctx, userID, roadmapID)
}

func (srv *roadmapService) ArchiveRoadmap(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.ArchiveRoadmap(ctx, userID, roadmapID)
}

// A restored technology re-occupies a slot in whichever goal it belongs to.
// An orphan (goal_id NULL) has no goal to fill, so it restores freely.
func (srv *roadmapService) RestoreRoadmap(ctx context.Context, userID, roadmapID int64) error {
	item, err := srv.repo.GetRoadmap(ctx, userID, roadmapID)
	if err != nil {
		return err
	}
	if item.GoalID != nil {
		count, err := srv.repo.CountRoadmapsInGoal(ctx, userID, *item.GoalID)
		if err != nil {
			return fmt.Errorf("restore roadmap: %w", err)
		}
		if count >= models.MaxRoadmapsPerGoal {
			return models.ErrRoadmapLimitReached
		}
	}
	return srv.repo.RestoreRoadmap(ctx, userID, roadmapID)
}

func (srv *roadmapService) DeleteRoadmapForever(ctx context.Context, userID, roadmapID int64) error {
	return srv.repo.DeleteRoadmapForever(ctx, userID, roadmapID)
}

func validRoadmapName(name string) bool {
	return !strings.Contains(name, "\n") && len(name) >= 2 && len(name) <= 60
}

func (srv *roadmapService) AddCardsFromText(ctx context.Context, userID, roadmapID int64, text string) (int, int, error) {
	if _, err := srv.repo.GetRoadmap(ctx, userID, roadmapID); err != nil {
		return 0, 0, err
	}

	cards := make([]models.RoadmapCardItem, 0)
	skipped := 0
	for _, line := range strings.Split(text, "\n") {
		card, ok := parseCardLine(line)
		if !ok {
			continue
		}
		if len([]rune(card.Text)) > models.MaxRoadmapCardTextLen {
			// Skipped rather than truncated: cutting a line in half loses
			// the user's text with no sign it happened, while a skip is
			// reported back as "(N line(s) skipped)".
			skipped++
			continue
		}
		cards = append(cards, card)
	}

	if len(cards) == 0 {
		return 0, skipped, models.ErrRoadmapNoCardsParsed
	}

	added, err := srv.repo.AddCards(ctx, roadmapID, cards)
	if err != nil {
		return 0, skipped, err
	}
	return added, skipped, nil
}

// People paste checklists straight out of notes, so a leading list marker
// would otherwise double up with the bullet the UI draws itself.
var cardLineMarkers = []string{"- ", "— ", "– ", "* ", "• ", "· "}

var cardKindTags = map[string]models.RoadmapCardKind{
	"#topic":   models.RoadmapCardTopic,
	"#article": models.RoadmapCardArticle,
	"#book":    models.RoadmapCardBook,
	"#lecture": models.RoadmapCardLecture,
}

var cardDifficultyTags = map[string]int{
	"!easy":   models.RoadmapCardEasy,
	"!mid":    models.RoadmapCardMedium,
	"!medium": models.RoadmapCardMedium,
	"!hard":   models.RoadmapCardHard,
}

// parseCardLine turns one pasted line into a card, pulling optional inline
// tags out of the text: "#book", "#article", "#lecture", "#topic" set the
// kind, "!easy"/"!mid"/"!hard" the difficulty. Tags are stripped from what
// gets stored, so "Kafka internals #book !hard" is saved as "Kafka
// internals".
//
// Tagging is optional by design — the point of the paste flow is dumping a
// plan in one go, so an untagged line still lands as a medium-difficulty
// topic. A line carrying a URL defaults to #article instead, since that's
// what a pasted link almost always is.
//
// ok is false for a line with nothing left after trimming.
func parseCardLine(line string) (models.RoadmapCardItem, bool) {
	line = strings.TrimSpace(line)
	for _, marker := range cardLineMarkers {
		if strings.HasPrefix(line, marker) {
			line = strings.TrimSpace(strings.TrimPrefix(line, marker))
			break
		}
	}
	if line == "" {
		return models.RoadmapCardItem{}, false
	}

	var (
		kind       models.RoadmapCardKind
		difficulty int
		kept       []string
	)
	for _, token := range strings.Fields(line) {
		lower := strings.ToLower(token)
		if k, ok := cardKindTags[lower]; ok {
			kind = k
			continue
		}
		if d, ok := cardDifficultyTags[lower]; ok {
			difficulty = d
			continue
		}
		kept = append(kept, token)
	}

	text := strings.Join(kept, " ")
	if text == "" {
		// The line was nothing but tags — no card to make out of it.
		return models.RoadmapCardItem{}, false
	}

	if kind == "" {
		kind = models.RoadmapCardTopic
		if strings.Contains(text, "http://") || strings.Contains(text, "https://") {
			kind = models.RoadmapCardArticle
		}
	}
	if difficulty == 0 {
		difficulty = models.RoadmapCardMedium
	}

	return models.RoadmapCardItem{Text: text, Kind: kind, Difficulty: difficulty}, true
}

func (srv *roadmapService) ListCards(ctx context.Context, userID, roadmapID int64) ([]models.RoadmapCardItem, error) {
	return srv.repo.ListCards(ctx, userID, roadmapID)
}

func (srv *roadmapService) ToggleCardDone(ctx context.Context, userID, cardID int64) (int64, error) {
	return srv.repo.ToggleCardDone(ctx, userID, cardID)
}

func (srv *roadmapService) CycleCardDifficulty(ctx context.Context, userID, cardID int64) (int64, error) {
	return srv.repo.CycleCardDifficulty(ctx, userID, cardID)
}

func (srv *roadmapService) DeleteCard(ctx context.Context, userID, cardID int64) error {
	return srv.repo.DeleteCard(ctx, userID, cardID)
}

// An already-running schedule on the same interval is kept as-is rather than
// restarted, so re-tapping the same interval doesn't push the next digest
// further away every time.
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

func (srv *roadmapService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop roadmap push: invalid userID")
	}
	return srv.repo.DisablePush(ctx, userID)
}

func (srv *roadmapService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.RoadmapDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return srv.repo.ListDueUsers(ctx, now.UTC(), limit)
}

func (srv *roadmapService) MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPushAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return srv.repo.SetNextPush(ctx, userID, nextPushAt)
}

func (srv *roadmapService) PickDigestCards(ctx context.Context, userID int64) ([]models.RoadmapDigestCard, error) {
	return srv.repo.PickDigestCards(ctx, userID, models.RoadmapDigestPerRoadmapCap, models.RoadmapDigestMaxCards)
}

func (srv *roadmapService) GetRoadmapStats(ctx context.Context, userID int64) (models.RoadmapStats, error) {
	goals, err := srv.repo.CountGoals(ctx, userID)
	if err != nil {
		return models.RoadmapStats{}, err
	}
	roadmaps, err := srv.repo.CountRoadmaps(ctx, userID)
	if err != nil {
		return models.RoadmapStats{}, err
	}
	total, done, err := srv.repo.CountCards(ctx, userID)
	if err != nil {
		return models.RoadmapStats{}, err
	}

	stats := models.RoadmapStats{
		TotalGoals:    goals,
		TotalRoadmaps: roadmaps,
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

func (srv *roadmapService) GetStatsDetail(ctx context.Context, userID int64) (models.RoadmapStatsDetail, error) {
	overall, err := srv.GetRoadmapStats(ctx, userID)
	if err != nil {
		return models.RoadmapStatsDetail{}, err
	}
	goals, err := srv.repo.ListGoals(ctx, userID, false)
	if err != nil {
		return models.RoadmapStatsDetail{}, err
	}
	roadmaps, err := srv.repo.GetRoadmapCardStats(ctx, userID)
	if err != nil {
		return models.RoadmapStatsDetail{}, err
	}

	return models.RoadmapStatsDetail{
		Overall:  overall,
		Goals:    goals,
		Roadmaps: roadmaps,
	}, nil
}
