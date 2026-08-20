package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
	"tracker-bot/pkg/apptime"
)

// LearningService contains word-learning use-cases: collections of word
// pairs, reviewed on a spaced-repetition schedule (SM-2-lite, see
// gradeSchedule) and delivered via periodic pushes (mirrors TimerService's
// Activate/Stop/ListDueUsers/MarkPromptSent shape).
type LearningService interface {
	CreateCollection(ctx context.Context, userID int64, name string) (int64, error)
	ListCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error)
	ListArchivedCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error)
	CollectionName(ctx context.Context, userID, collectionID int64) (string, error)
	RenameCollection(ctx context.Context, userID, collectionID int64, newName string) error
	ToggleCollectionActive(ctx context.Context, userID, collectionID int64) error
	ArchiveCollection(ctx context.Context, userID, collectionID int64) error
	RestoreCollection(ctx context.Context, userID, collectionID int64) error
	DeleteCollectionForever(ctx context.Context, userID, collectionID int64) error

	// AddWordsFromText parses free-text "word - translation" lines (one per
	// line) and appends them to a collection. Returns how many lines parsed
	// into words versus were skipped as malformed.
	AddWordsFromText(ctx context.Context, userID, collectionID int64, text string) (added, skipped int, err error)
	ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error)
	DeleteWord(ctx context.Context, userID, wordID int64) error

	// Review delivery + grading.
	PickDueWord(ctx context.Context, userID int64) (*models.LearningDueWord, error)
	// PeekWord loads a word's display text (collection name, term,
	// translation) without touching its SRS state — used to render the
	// "reveal" step of a review card before the user grades it.
	PeekWord(ctx context.Context, userID, wordID int64) (collectionName, term, translation string, err error)
	// GradeAnswer applies an Anki-style Again/Hard/Good/Easy grade to a
	// word's schedule (see gradeSchedule) and records the review.
	GradeAnswer(ctx context.Context, userID, wordID int64, grade models.LearningGrade) (nextIntervalDays int, learned bool, err error)
	// ListReviewsOnDay returns every review the user answered within
	// [from, to) — used by the tracking heatmap's day drill-down.
	ListReviewsOnDay(ctx context.Context, userID int64, from, to time.Time) ([]models.LearningReviewEntry, error)

	// Push scheduling.
	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error)
	MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error

	GetLearningStats(ctx context.Context, userID int64, loc *time.Location) (models.LearningStats, error)
	// GetStatsDetail returns the full breakdown behind the "📈 Statistics"
	// screen: overall numbers, per-collection counts, and answer accuracy.
	GetStatsDetail(ctx context.Context, userID int64, loc *time.Location) (models.LearningStatsDetail, error)
}

type learningService struct {
	repo repo.LearningRepository
}

func NewLearningService(repo repo.LearningRepository) LearningService {
	return &learningService{
		repo: repo,
	}
}

// CreateCollection validates and stores a new named collection.
func (srv *learningService) CreateCollection(ctx context.Context, userID int64, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 {
		return 0, fmt.Errorf("create collection: invalid userID")
	}
	if len(name) < 2 {
		return 0, fmt.Errorf("create collection: name too short")
	}
	return srv.repo.CreateCollection(ctx, userID, name)
}

// ListCollections returns a user's active (non-archived) collections.
func (srv *learningService) ListCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error) {
	return srv.repo.ListCollections(ctx, userID, false)
}

// ListArchivedCollections returns a user's archived collections.
func (srv *learningService) ListArchivedCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error) {
	return srv.repo.ListCollections(ctx, userID, true)
}

// CollectionName resolves a collection's display name.
func (srv *learningService) CollectionName(ctx context.Context, userID, collectionID int64) (string, error) {
	return srv.repo.GetCollectionName(ctx, userID, collectionID)
}

// RenameCollection validates and applies a new name for a collection.
func (srv *learningService) RenameCollection(ctx context.Context, userID, collectionID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if strings.Contains(newName, "\n") || len(newName) < 2 || len(newName) > 60 {
		return models.ErrLearningInvalidName
	}
	return srv.repo.RenameCollection(ctx, userID, collectionID, newName)
}

// ToggleCollectionActive flips whether a collection participates in review
// pushes.
func (srv *learningService) ToggleCollectionActive(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.ToggleCollectionActive(ctx, userID, collectionID)
}

// ArchiveCollection moves a collection to the archive.
func (srv *learningService) ArchiveCollection(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.ArchiveCollection(ctx, userID, collectionID)
}

// RestoreCollection moves an archived collection back to the active list.
func (srv *learningService) RestoreCollection(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.RestoreCollection(ctx, userID, collectionID)
}

// DeleteCollectionForever permanently removes a collection and its words.
func (srv *learningService) DeleteCollectionForever(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.DeleteCollectionForever(ctx, userID, collectionID)
}

// AddWordsFromText parses one "word - translation" pair per line (also
// accepting "–", "—" or ":" as the separator) and appends valid pairs to the
// collection. Blank lines and lines with no valid separator/either side
// empty are counted as skipped rather than erroring the whole batch.
func (srv *learningService) AddWordsFromText(ctx context.Context, userID, collectionID int64, text string) (int, int, error) {
	if _, err := srv.repo.GetCollectionName(ctx, userID, collectionID); err != nil {
		return 0, 0, err
	}

	pairs := make([]models.LearningWordItem, 0)
	skipped := 0
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		term, translation, ok := parseWordLine(line)
		if !ok {
			skipped++
			continue
		}
		pairs = append(pairs, models.LearningWordItem{Term: term, Translation: translation})
	}

	if len(pairs) == 0 {
		if skipped > 0 {
			return 0, skipped, models.ErrLearningNoWordsParsed
		}
		return 0, 0, models.ErrLearningNoWordsParsed
	}

	added, err := srv.repo.AddWords(ctx, collectionID, pairs)
	if err != nil {
		return 0, skipped, err
	}
	return added, skipped, nil
}

// wordLineSeparators are tried in order; the first one present in the line
// splits it into term/translation. The bare "–"/"—" entries matter beyond
// the padded " – "/" — " ones: mobile "smart punctuation" autocorrect often
// turns a typed "-" into an en/em dash *without* re-adding the surrounding
// spaces (e.g. "apple—яблоко"), which would otherwise fail to parse at all.
var wordLineSeparators = []string{" - ", " – ", " — ", ":", "–", "—", "-"}

// parseWordLine splits one line into (term, translation). ok is false when
// no separator is found or either side is empty after trimming.
func parseWordLine(line string) (string, string, bool) {
	for _, sep := range wordLineSeparators {
		idx := strings.Index(line, sep)
		if idx <= 0 {
			continue
		}
		term := strings.TrimSpace(line[:idx])
		translation := strings.TrimSpace(line[idx+len(sep):])
		if term == "" || translation == "" {
			continue
		}
		return term, translation, true
	}
	return "", "", false
}

// ListWords returns every word in one collection.
func (srv *learningService) ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error) {
	return srv.repo.ListWords(ctx, userID, collectionID)
}

// DeleteWord removes one word pair.
func (srv *learningService) DeleteWord(ctx context.Context, userID, wordID int64) error {
	return srv.repo.DeleteWord(ctx, userID, wordID)
}

// PickDueWord returns the most-overdue due word from the user's active
// collections, or nil if none is due right now.
func (srv *learningService) PickDueWord(ctx context.Context, userID int64) (*models.LearningDueWord, error) {
	return srv.repo.PickDueWord(ctx, userID, time.Now().UTC())
}

// PeekWord loads a word's display text without touching its SRS state.
func (srv *learningService) PeekWord(ctx context.Context, userID, wordID int64) (string, string, string, error) {
	collectionName, term, translation, _, _, _, err := srv.repo.GetWordForGrading(ctx, userID, wordID)
	if err != nil {
		return "", "", "", err
	}
	return collectionName, term, translation, nil
}

// GradeAnswer applies an Anki-style Again/Hard/Good/Easy update (see
// gradeSchedule) and records the review for stats/streak. A word counts as
// "correct" for accuracy stats whenever the grade isn't Again — Hard still
// means they recalled it, just with effort.
func (srv *learningService) GradeAnswer(ctx context.Context, userID, wordID int64, grade models.LearningGrade) (int, bool, error) {
	_, _, _, easeFactor, intervalDays, repetitions, err := srv.repo.GetWordForGrading(ctx, userID, wordID)
	if err != nil {
		return 0, false, err
	}

	newEase, newInterval, newReps, learned := gradeSchedule(easeFactor, intervalDays, repetitions, grade)
	now := time.Now().UTC()
	nextReviewAt := now.Add(time.Duration(newInterval) * 24 * time.Hour)

	if err := srv.repo.UpdateWordSchedule(ctx, wordID, newEase, newInterval, newReps, nextReviewAt, learned); err != nil {
		return 0, false, err
	}
	correct := grade != models.LearningGradeAgain
	if err := srv.repo.RecordReview(ctx, wordID, userID, correct, now); err != nil {
		return 0, false, err
	}
	return newInterval, learned, nil
}

// ListReviewsOnDay returns every review the user answered within [from, to).
func (srv *learningService) ListReviewsOnDay(ctx context.Context, userID int64, from, to time.Time) ([]models.LearningReviewEntry, error) {
	return srv.repo.ListReviewsOnDay(ctx, userID, from, to)
}

// gradeSchedule is an Anki-style SM-2 variant with four grades instead of a
// plain correct/incorrect:
//
//   - Again: forgot it — interval resets to 1 day, repetitions reset to 0,
//     ease drops. Never counts as "learned".
//   - Hard: recalled, but it was a struggle — interval grows slowly (~1.2x),
//     ease drops slightly.
//   - Good: recalled comfortably — the original growth curve (1 -> 6 ->
//     interval*ease days), ease unchanged. This is the old binary
//     "correct" path.
//   - Easy: trivially easy — skips straight to a longer interval (4 days
//     for a brand-new word) and grows faster (1.3x bonus), ease rises.
//
// Ease factor stays within SM-2's usual [1.3, 2.5] bounds throughout. A
// word graduates to "learned" once its interval reaches 3 weeks via
// anything but Again.
func gradeSchedule(easeFactor float32, intervalDays, repetitions int, grade models.LearningGrade) (newEase float32, newInterval, newReps int, learned bool) {
	const (
		minEase          = 1.3
		maxEase          = 2.5
		learnedThreshold = 21 // days
	)

	switch grade {
	case models.LearningGradeAgain:
		return maxf32(easeFactor-0.20, minEase), 1, 0, false

	case models.LearningGradeHard:
		newReps = repetitions + 1
		newInterval = maxInt(intervalDays+1, roundf32(float32(intervalDays)*1.2))
		if repetitions == 0 {
			newInterval = 1
		}
		newEase = maxf32(easeFactor-0.15, minEase)

	case models.LearningGradeEasy:
		newReps = repetitions + 1
		if repetitions == 0 {
			newInterval = 4
		} else {
			newInterval = maxInt(intervalDays+1, roundf32(float32(intervalDays)*easeFactor*1.3))
		}
		newEase = minf32(easeFactor+0.15, maxEase)

	default: // models.LearningGradeGood
		newReps = repetitions + 1
		switch newReps {
		case 1:
			newInterval = 1
		case 2:
			newInterval = 6
		default:
			newInterval = maxInt(intervalDays+1, roundf32(float32(intervalDays)*easeFactor))
		}
		newEase = easeFactor
	}

	learned = newInterval >= learnedThreshold
	return newEase, newInterval, newReps, learned
}

func roundf32(v float32) int {
	return int(v + 0.5)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minf32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// Activate enables review pushes and schedules the next one. Mirrors
// TimerService.Activate: an already-running schedule with the same interval
// is kept as-is rather than restarted.
func (srv *learningService) Activate(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 {
		return fmt.Errorf("activate learning push: invalid userID")
	}
	if intervalMin <= 0 || intervalMin > 1440 {
		return models.ErrLearningInvalidInterval
	}

	now := time.Now().UTC()
	nextPushAt := now.Add(time.Duration(intervalMin) * time.Minute)

	curInterval, curNextPushAt, active, err := srv.repo.GetPushSettings(ctx, userID)
	if err != nil {
		return fmt.Errorf("activate learning push: %w", err)
	}
	if active && curInterval == intervalMin && curNextPushAt.After(now) {
		nextPushAt = curNextPushAt
	}

	return srv.repo.UpsertPushInterval(ctx, userID, intervalMin, nextPushAt)
}

// Stop disables review pushes for user.
func (srv *learningService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop learning push: invalid userID")
	}
	return srv.repo.DisablePush(ctx, userID)
}

// ListDueUsers returns users that should receive a review push now.
func (srv *learningService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return srv.repo.ListDueUsers(ctx, now.UTC(), limit)
}

// MarkPushSent moves next push time forward by interval.
func (srv *learningService) MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPushAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return srv.repo.SetNextPush(ctx, userID, nextPushAt)
}

// GetLearningStats aggregates the learning dashboard's numbers.
func (srv *learningService) GetLearningStats(ctx context.Context, userID int64, loc *time.Location) (models.LearningStats, error) {
	total, dueToday, learned, err := srv.repo.CountWords(ctx, userID)
	if err != nil {
		return models.LearningStats{}, err
	}

	streak, err := srv.computeStreak(ctx, userID, apptime.NowIn(loc))
	if err != nil {
		return models.LearningStats{}, err
	}

	stats := models.LearningStats{
		TotalWords:    total,
		DueTodayWords: dueToday,
		LearnedWords:  learned,
		StreakDays:    streak,
	}

	intervalMin, nextPushAt, active, err := srv.repo.GetPushSettings(ctx, userID)
	if err != nil {
		return models.LearningStats{}, err
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
// plus a per-collection table and answer accuracy.
func (srv *learningService) GetStatsDetail(ctx context.Context, userID int64, loc *time.Location) (models.LearningStatsDetail, error) {
	overall, err := srv.GetLearningStats(ctx, userID, loc)
	if err != nil {
		return models.LearningStatsDetail{}, err
	}

	collections, err := srv.repo.GetCollectionStats(ctx, userID)
	if err != nil {
		return models.LearningStatsDetail{}, err
	}

	correct, total, err := srv.repo.GetAccuracy(ctx, userID)
	if err != nil {
		return models.LearningStatsDetail{}, err
	}

	return models.LearningStatsDetail{
		Overall:        overall,
		Collections:    collections,
		ReviewsTotal:   total,
		ReviewsCorrect: correct,
	}, nil
}

// computeStreak counts consecutive calendar days, in now's timezone, (ending
// today or yesterday) with at least one recorded review. now is expected to
// already be resolved to the user's own timezone (see apptime.NowIn), since
// its Location determines every day boundary below.
func (srv *learningService) computeStreak(ctx context.Context, userID int64, now time.Time) (int, error) {
	loc := now.Location()
	since := now.AddDate(0, 0, -90)
	dates, err := srv.repo.ListReviewDates(ctx, userID, since)
	if err != nil {
		return 0, err
	}
	if len(dates) == 0 {
		return 0, nil
	}

	today := truncDayIn(now, loc)
	cursor := today
	if !truncDayIn(dates[0], loc).Equal(today) {
		// Streak isn't "live" today — allow it to still count if the most
		// recent review was yesterday, otherwise it's broken.
		yesterday := today.AddDate(0, 0, -1)
		if !truncDayIn(dates[0], loc).Equal(yesterday) {
			return 0, nil
		}
		cursor = yesterday
	}

	streak := 0
	idx := 0
	for idx < len(dates) {
		d := truncDayIn(dates[idx], loc)
		if d.Equal(cursor) {
			streak++
			cursor = cursor.AddDate(0, 0, -1)
			idx++
			continue
		}
		if d.Before(cursor) {
			break
		}
		idx++
	}
	return streak, nil
}

// truncDayIn truncates t to midnight of its calendar day in loc.
func truncDayIn(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// truncDay truncates t to midnight UTC of its calendar day — used by
// challenge.go where days are already plain calendar dates (DATE columns,
// parsed in a fixed location), not instants needing timezone conversion.
func truncDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
