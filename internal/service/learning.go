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

	AddWordsFromText(ctx context.Context, userID, collectionID int64, text string) (added, skipped int, err error)
	ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error)
	DeleteWord(ctx context.Context, userID, wordID int64) error

	PickDueWord(ctx context.Context, userID int64) (*models.LearningDueWord, error)
	PeekWord(ctx context.Context, userID, wordID int64) (collectionName, term, translation string, err error)
	PreviewGradeDelays(ctx context.Context, userID, wordID int64) (again, hard, good, easy time.Duration, err error)
	// GradeAnswer: nextReviewAt may be minutes away (Again/new-word Hard), not always a full day.
	GradeAnswer(ctx context.Context, userID, wordID int64, grade models.LearningGrade) (nextReviewAt time.Time, learned bool, err error)
	ListReviewsOnDay(ctx context.Context, userID int64, from, to time.Time) ([]models.LearningReviewEntry, error)

	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error)
	MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error

	GetLearningStats(ctx context.Context, userID int64, loc *time.Location) (models.LearningStats, error)
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

func (srv *learningService) ListCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error) {
	return srv.repo.ListCollections(ctx, userID, false)
}

func (srv *learningService) ListArchivedCollections(ctx context.Context, userID int64) ([]models.LearningCollectionItem, error) {
	return srv.repo.ListCollections(ctx, userID, true)
}

func (srv *learningService) CollectionName(ctx context.Context, userID, collectionID int64) (string, error) {
	return srv.repo.GetCollectionName(ctx, userID, collectionID)
}

func (srv *learningService) RenameCollection(ctx context.Context, userID, collectionID int64, newName string) error {
	newName = strings.TrimSpace(newName)
	if strings.Contains(newName, "\n") || len(newName) < 2 || len(newName) > 60 {
		return models.ErrLearningInvalidName
	}
	return srv.repo.RenameCollection(ctx, userID, collectionID, newName)
}

func (srv *learningService) ToggleCollectionActive(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.ToggleCollectionActive(ctx, userID, collectionID)
}

func (srv *learningService) ArchiveCollection(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.ArchiveCollection(ctx, userID, collectionID)
}

func (srv *learningService) RestoreCollection(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.RestoreCollection(ctx, userID, collectionID)
}

func (srv *learningService) DeleteCollectionForever(ctx context.Context, userID, collectionID int64) error {
	return srv.repo.DeleteCollectionForever(ctx, userID, collectionID)
}

// AddWordsFromText: malformed/blank lines are counted as skipped, not treated as a batch error.
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

// bare "–"/"—" entries matter: mobile autocorrect turns "-" into a dash with no surrounding spaces.
var wordLineSeparators = []string{" - ", " – ", " — ", ":", "–", "—", "-"}

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

func (srv *learningService) ListWords(ctx context.Context, userID, collectionID int64) ([]models.LearningWordItem, error) {
	return srv.repo.ListWords(ctx, userID, collectionID)
}

func (srv *learningService) DeleteWord(ctx context.Context, userID, wordID int64) error {
	return srv.repo.DeleteWord(ctx, userID, wordID)
}

// PickDueWord returns nil (not an error) when no word is due right now.
func (srv *learningService) PickDueWord(ctx context.Context, userID int64) (*models.LearningDueWord, error) {
	return srv.repo.PickDueWord(ctx, userID, time.Now().UTC())
}

func (srv *learningService) PeekWord(ctx context.Context, userID, wordID int64) (string, string, string, error) {
	collectionName, term, translation, _, _, _, err := srv.repo.GetWordForGrading(ctx, userID, wordID)
	if err != nil {
		return "", "", "", err
	}
	return collectionName, term, translation, nil
}

func (srv *learningService) PreviewGradeDelays(ctx context.Context, userID, wordID int64) (again, hard, good, easy time.Duration, err error) {
	_, _, _, easeFactor, intervalDays, repetitions, err := srv.repo.GetWordForGrading(ctx, userID, wordID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	for _, g := range []struct {
		grade models.LearningGrade
		out   *time.Duration
	}{
		{models.LearningGradeAgain, &again},
		{models.LearningGradeHard, &hard},
		{models.LearningGradeGood, &good},
		{models.LearningGradeEasy, &easy},
	} {
		_, newInterval, _, _ := gradeSchedule(easeFactor, intervalDays, repetitions, g.grade)
		*g.out = reviewDelay(g.grade, repetitions, newInterval)
	}
	return again, hard, good, easy, nil
}

// GradeAnswer: for accuracy stats, only Again counts as incorrect — Hard still counts as correct.
func (srv *learningService) GradeAnswer(ctx context.Context, userID, wordID int64, grade models.LearningGrade) (time.Time, bool, error) {
	_, _, _, easeFactor, intervalDays, repetitions, err := srv.repo.GetWordForGrading(ctx, userID, wordID)
	if err != nil {
		return time.Time{}, false, err
	}

	newEase, newInterval, newReps, learned := gradeSchedule(easeFactor, intervalDays, repetitions, grade)
	now := time.Now().UTC()
	nextReviewAt := now.Add(reviewDelay(grade, repetitions, newInterval))

	if err := srv.repo.UpdateWordSchedule(ctx, wordID, newEase, newInterval, newReps, nextReviewAt, learned); err != nil {
		return time.Time{}, false, err
	}
	correct := grade != models.LearningGradeAgain
	if err := srv.repo.RecordReview(ctx, wordID, userID, correct, now); err != nil {
		return time.Time{}, false, err
	}
	return nextReviewAt, learned, nil
}

func reviewDelay(grade models.LearningGrade, oldRepetitions, newIntervalDays int) time.Duration {
	switch grade {
	case models.LearningGradeAgain:
		return 10 * time.Minute
	case models.LearningGradeHard:
		if oldRepetitions == 0 {
			return 15 * time.Minute
		}
	}
	return time.Duration(newIntervalDays) * 24 * time.Hour
}

func (srv *learningService) ListReviewsOnDay(ctx context.Context, userID int64, from, to time.Time) ([]models.LearningReviewEntry, error) {
	return srv.repo.ListReviewsOnDay(ctx, userID, from, to)
}

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

// Activate: an already-running schedule with the same interval is kept as-is, not restarted.
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

func (srv *learningService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop learning push: invalid userID")
	}
	return srv.repo.DisablePush(ctx, userID)
}

func (srv *learningService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.LearningDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return srv.repo.ListDueUsers(ctx, now.UTC(), limit)
}

func (srv *learningService) MarkPushSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPushAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return srv.repo.SetNextPush(ctx, userID, nextPushAt)
}

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

// computeStreak: now must already be in the user's timezone — its Location sets every day boundary.
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

func truncDayIn(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// truncDay: UTC-midnight truncation for challenge.go's plain DATE columns, not tz-aware instants.
func truncDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
