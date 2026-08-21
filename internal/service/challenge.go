package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

const challengePushHour = 21

// challengeTrendWindow is how many days a day-detail screen's trend strip
// looks back, ending at (and including) the tapped day.
const challengeTrendWindow = 14

// mirrors TimerService's shape, but "interval" here is always "same wall-clock hour tomorrow"
type ChallengeService interface {
	CreateChallenge(ctx context.Context, userID int64, name string, startDate, endDate time.Time, loc *time.Location) (int64, error)
	ListChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error)
	ListArchivedChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error)
	GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error)
	ArchiveChallenge(ctx context.Context, userID, challengeID int64) error
	RestoreChallenge(ctx context.Context, userID, challengeID int64, loc *time.Location) error
	DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error

	ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error)
	GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error)
	MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, done bool) error
	// GetDayDetail is the day-square tap's view model: that day's status
	// plus current/best streak (as of today) and a trend window ending at
	// day. today and day need not be truncated by the caller.
	GetDayDetail(ctx context.Context, userID, challengeID int64, day, today time.Time) (models.ChallengeDayDetail, error)

	ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error)
	AdvancePush(ctx context.Context, challengeID int64, firedAt time.Time, endDate time.Time, loc *time.Location) error
}

type challengeService struct {
	repo repo.ChallengeRepository
}

func NewChallengeService(repo repo.ChallengeRepository) ChallengeService {
	return &challengeService{repo: repo}
}

// also pre-populates challenge_days for the whole range (see repo.Create)
func (s *challengeService) CreateChallenge(ctx context.Context, userID int64, name string, startDate, endDate time.Time, loc *time.Location) (int64, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 {
		return 0, fmt.Errorf("create challenge: invalid userID")
	}
	if strings.Contains(name, "\n") || len(name) < 2 || len(name) > 60 {
		return 0, models.ErrChallengeInvalidName
	}
	startDate = truncDay(startDate)
	endDate = truncDay(endDate)
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalDays < 1 || totalDays > 100 {
		return 0, models.ErrChallengeInvalidRange
	}

	id, err := s.repo.Create(ctx, userID, name, startDate, endDate)
	if err != nil {
		return 0, err
	}

	if nextPush, ok := nextChallengeFireTime(resolveLoc(loc), time.Now(), startDate, endDate); ok {
		if err := s.repo.UpsertPushSchedule(ctx, id, nextPush); err != nil {
			return id, err
		}
	}
	return id, nil
}

func (s *challengeService) ListChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error) {
	return s.repo.ListChallenges(ctx, userID, false)
}

func (s *challengeService) ListArchivedChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error) {
	return s.repo.ListChallenges(ctx, userID, true)
}

func (s *challengeService) GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error) {
	return s.repo.GetChallenge(ctx, userID, challengeID)
}

// stopping the push schedule is handled atomically by the repo, not here
func (s *challengeService) ArchiveChallenge(ctx context.Context, userID, challengeID int64) error {
	return s.repo.ArchiveChallenge(ctx, userID, challengeID)
}

func (s *challengeService) RestoreChallenge(ctx context.Context, userID, challengeID int64, loc *time.Location) error {
	if err := s.repo.RestoreChallenge(ctx, userID, challengeID); err != nil {
		return err
	}
	item, err := s.repo.GetChallenge(ctx, userID, challengeID)
	if err != nil {
		return err
	}
	if nextPush, ok := nextChallengeFireTime(resolveLoc(loc), time.Now(), item.StartDate, item.EndDate); ok {
		return s.repo.UpsertPushSchedule(ctx, challengeID, nextPush)
	}
	return nil
}

func (s *challengeService) DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error {
	return s.repo.DeleteChallengeForever(ctx, userID, challengeID)
}

func (s *challengeService) ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error) {
	return s.repo.ListDays(ctx, userID, challengeID)
}

func (s *challengeService) GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error) {
	return s.repo.GetDayStatus(ctx, userID, challengeID, truncDay(day))
}

func (s *challengeService) MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, done bool) error {
	status := models.ChallengeDaySkipped
	if done {
		status = models.ChallengeDayDone
	}
	return s.repo.MarkDay(ctx, userID, challengeID, truncDay(day), status)
}

func (s *challengeService) GetDayDetail(ctx context.Context, userID, challengeID int64, day, today time.Time) (models.ChallengeDayDetail, error) {
	day = truncDay(day)
	status, err := s.repo.GetDayStatus(ctx, userID, challengeID, day)
	if err != nil {
		return models.ChallengeDayDetail{}, err
	}
	days, err := s.repo.ListDays(ctx, userID, challengeID)
	if err != nil {
		return models.ChallengeDayDetail{}, err
	}
	current, best := challengeStreaks(days, truncDay(today))
	return models.ChallengeDayDetail{
		Day:           day,
		Status:        status,
		CurrentStreak: current,
		BestStreak:    best,
		Trend:         challengeTrend(days, day, challengeTrendWindow),
	}, nil
}

// challengeStreaks computes the current and best consecutive-"done" runs
// from days (any order, plain UTC calendar dates). today should already be
// truncDay-ed.
//
// A day with no explicit status yet — still pending, or past the last
// recorded entry — gets the benefit of the doubt and the walk starts from
// yesterday instead, so the streak doesn't zero out before the user has had
// a chance to check in tonight. An explicit skip breaks it immediately,
// same day, since that's a deliberate "no" rather than "not yet decided".
func challengeStreaks(days []models.ChallengeDay, today time.Time) (current, best int) {
	byDate := make(map[time.Time]models.ChallengeDayStatus, len(days))
	for _, d := range days {
		byDate[truncDay(d.Date)] = d.Status
	}

	cursor := today
	if byDate[cursor] != models.ChallengeDayDone && byDate[cursor] != models.ChallengeDaySkipped {
		cursor = cursor.AddDate(0, 0, -1)
	}
	for byDate[cursor] == models.ChallengeDayDone {
		current++
		cursor = cursor.AddDate(0, 0, -1)
	}

	run := 0
	for _, d := range days {
		if d.Status == models.ChallengeDayDone {
			run++
			if run > best {
				best = run
			}
		} else {
			run = 0
		}
	}
	return current, best
}

// challengeTrend returns up to window days' statuses ending at (and
// including) day, oldest first, clipped to the earliest day present in
// days so a young challenge doesn't report phantom pre-start days.
func challengeTrend(days []models.ChallengeDay, day time.Time, window int) []models.ChallengeDayStatus {
	if len(days) == 0 {
		return nil
	}
	byDate := make(map[time.Time]models.ChallengeDayStatus, len(days))
	earliest := truncDay(days[0].Date)
	for _, d := range days {
		dt := truncDay(d.Date)
		byDate[dt] = d.Status
		if dt.Before(earliest) {
			earliest = dt
		}
	}

	start := day.AddDate(0, 0, -(window - 1))
	if start.Before(earliest) {
		start = earliest
	}

	out := make([]models.ChallengeDayStatus, 0, window)
	for d := start; !d.After(day); d = d.AddDate(0, 0, 1) {
		out = append(out, byDate[d])
	}
	return out
}

func (s *challengeService) ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListDueChallenges(ctx, now.UTC(), limit)
}

func (s *challengeService) AdvancePush(ctx context.Context, challengeID int64, firedAt time.Time, endDate time.Time, loc *time.Location) error {
	next, ok := nextDailyFireTime(resolveLoc(loc), firedAt, endDate)
	if !ok {
		return s.repo.ClearPushSchedule(ctx, challengeID)
	}
	return s.repo.UpsertPushSchedule(ctx, challengeID, next)
}

// next evening-push instant on/after from, clamped into [startDate,endDate]; ok=false once range has elapsed
func nextChallengeFireTime(loc *time.Location, from time.Time, startDate, endDate time.Time) (time.Time, bool) {
	from = from.In(loc)
	candidate := time.Date(from.Year(), from.Month(), from.Day(), challengePushHour, 0, 0, 0, loc)
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), challengePushHour, 0, 0, 0, loc)
	if start.After(candidate) {
		candidate = start
	}
	if candidate.Before(from) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), challengePushHour, 0, 0, 0, loc)
	if candidate.After(end) {
		return time.Time{}, false
	}
	return candidate, true
}

func nextDailyFireTime(loc *time.Location, firedAt time.Time, endDate time.Time) (time.Time, bool) {
	next := firedAt.In(loc).AddDate(0, 0, 1)
	next = time.Date(next.Year(), next.Month(), next.Day(), challengePushHour, 0, 0, 0, loc)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), challengePushHour, 0, 0, 0, loc)
	if next.After(end) {
		return time.Time{}, false
	}
	return next, true
}
