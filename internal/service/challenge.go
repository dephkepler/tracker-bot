package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// challengePushHour is the fixed local wall-clock hour each day's evening
// "did you do it?" push fires at.
const challengePushHour = 21

// ChallengeService contains "day-range plan with a square per day"
// use-cases: create a challenge, mark days done/skipped, and the daily
// evening push schedule (mirrors TimerService's shape, but the "interval"
// here is always "the same wall-clock hour tomorrow").
type ChallengeService interface {
	// CreateChallenge validates and creates a challenge, then schedules its
	// first evening push.
	CreateChallenge(ctx context.Context, userID int64, name string, startDate, endDate time.Time, loc *time.Location) (int64, error)
	ListChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error)
	ListArchivedChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error)
	GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error)
	ArchiveChallenge(ctx context.Context, userID, challengeID int64) error
	// RestoreChallenge moves a challenge back to the active list and
	// re-schedules its push if any days remain.
	RestoreChallenge(ctx context.Context, userID, challengeID int64, loc *time.Location) error
	DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error

	ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error)
	GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error)
	MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, done bool) error

	ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error)
	// AdvancePush schedules the next evening push (tomorrow, same hour),
	// or clears the schedule once the challenge's range has ended.
	AdvancePush(ctx context.Context, challengeID int64, firedAt time.Time, endDate time.Time, loc *time.Location) error
}

type challengeService struct {
	repo repo.ChallengeRepository
}

// NewChallengeService creates challenge service.
func NewChallengeService(repo repo.ChallengeRepository) ChallengeService {
	return &challengeService{repo: repo}
}

// CreateChallenge validates name/range and stores a new challenge, with
// challenge_days pre-populated for the whole range (see repo.Create).
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

// ListChallenges returns a user's active (non-archived) challenges.
func (s *challengeService) ListChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error) {
	return s.repo.ListChallenges(ctx, userID, false)
}

// ListArchivedChallenges returns a user's archived challenges.
func (s *challengeService) ListArchivedChallenges(ctx context.Context, userID int64) ([]models.ChallengeItem, error) {
	return s.repo.ListChallenges(ctx, userID, true)
}

// GetChallenge loads one challenge's summary.
func (s *challengeService) GetChallenge(ctx context.Context, userID, challengeID int64) (models.ChallengeItem, error) {
	return s.repo.GetChallenge(ctx, userID, challengeID)
}

// ArchiveChallenge moves a challenge to the archive and stops its push
// (handled by the repo in the same statement).
func (s *challengeService) ArchiveChallenge(ctx context.Context, userID, challengeID int64) error {
	return s.repo.ArchiveChallenge(ctx, userID, challengeID)
}

// RestoreChallenge moves an archived challenge back to the active list and
// re-schedules its evening push if any days remain in its range.
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

// DeleteChallengeForever permanently removes a challenge and its days.
func (s *challengeService) DeleteChallengeForever(ctx context.Context, userID, challengeID int64) error {
	return s.repo.DeleteChallengeForever(ctx, userID, challengeID)
}

// ListDays returns every day in a challenge's range.
func (s *challengeService) ListDays(ctx context.Context, userID, challengeID int64) ([]models.ChallengeDay, error) {
	return s.repo.ListDays(ctx, userID, challengeID)
}

// GetDayStatus returns one day's current status.
func (s *challengeService) GetDayStatus(ctx context.Context, userID, challengeID int64, day time.Time) (models.ChallengeDayStatus, error) {
	return s.repo.GetDayStatus(ctx, userID, challengeID, truncDay(day))
}

// MarkDay sets one day done or skipped.
func (s *challengeService) MarkDay(ctx context.Context, userID, challengeID int64, day time.Time, done bool) error {
	status := models.ChallengeDaySkipped
	if done {
		status = models.ChallengeDayDone
	}
	return s.repo.MarkDay(ctx, userID, challengeID, truncDay(day), status)
}

// ListDueChallenges returns challenges due for their evening push now.
func (s *challengeService) ListDueChallenges(ctx context.Context, now time.Time, limit int) ([]models.ChallengeDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListDueChallenges(ctx, now.UTC(), limit)
}

// AdvancePush schedules tomorrow's push at the same wall-clock hour, or
// clears the schedule once the challenge's range has ended.
func (s *challengeService) AdvancePush(ctx context.Context, challengeID int64, firedAt time.Time, endDate time.Time, loc *time.Location) error {
	next, ok := nextDailyFireTime(resolveLoc(loc), firedAt, endDate)
	if !ok {
		return s.repo.ClearPushSchedule(ctx, challengeID)
	}
	return s.repo.UpsertPushSchedule(ctx, challengeID, next)
}

// nextChallengeFireTime returns the next evening-push instant at or after
// from, clamped into [startDate, endDate] at challengePushHour local time.
// ok is false once the challenge's range has already fully elapsed.
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

// nextDailyFireTime returns the next day's push at challengePushHour local
// time after firedAt, or ok=false once past endDate.
func nextDailyFireTime(loc *time.Location, firedAt time.Time, endDate time.Time) (time.Time, bool) {
	next := firedAt.In(loc).AddDate(0, 0, 1)
	next = time.Date(next.Year(), next.Month(), next.Day(), challengePushHour, 0, 0, 0, loc)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), challengePushHour, 0, 0, 0, loc)
	if next.After(end) {
		return time.Time{}, false
	}
	return next, true
}
