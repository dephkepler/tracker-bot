package service

import (
	"context"
	"fmt"
	"sort"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

// TimerService contains timer-related use-cases.
type TimerService interface {
	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error)
	MarkPromptSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error
	RecordPromptAnswer(ctx context.Context, userID, activityID int64) error
	RecordPromptAnswerWithInterval(ctx context.Context, userID, activityID int64, intervalMin int) error

	// AddCustomInterval saves a user-defined timer interval (in addition to
	// the built-in 15/30 min choices) so it shows up in the picker next to
	// them. Adding an interval that already exists is a no-op.
	AddCustomInterval(ctx context.Context, userID int64, intervalMin int) error
	// ListCustomIntervals returns the user's custom intervals, ascending.
	ListCustomIntervals(ctx context.Context, userID int64) ([]int, error)
	// RemoveCustomInterval deletes one previously added custom interval.
	RemoveCustomInterval(ctx context.Context, userID int64, intervalMin int) error
}

type timerService struct {
	timerRepo       repo.TimerRepository
	sessionRepo     repo.SessionRepository
	customTimerRepo repo.CustomTimerRepository
}

// NewTimerService creates timer service.
func NewTimerService(timerRepo repo.TimerRepository, sessionRepo repo.SessionRepository, customTimerRepo repo.CustomTimerRepository) TimerService {
	return &timerService{
		timerRepo:       timerRepo,
		sessionRepo:     sessionRepo,
		customTimerRepo: customTimerRepo,
	}
}

// Activate enables timer and schedules next prompt. If a timer is already
// running with the same interval, the existing schedule is kept as-is
// instead of being restarted — this lets a newly selected activity join the
// already-running countdown (e.g. adding a 4th activity mid-day) rather than
// pushing every previously scheduled prompt back by a full interval.
func (s *timerService) Activate(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 {
		return fmt.Errorf("activate timer: invalid userID")
	}
	if intervalMin <= 0 {
		return fmt.Errorf("activate timer: invalid interval")
	}

	now := time.Now().UTC()
	nextPingAt := now.Add(time.Duration(intervalMin) * time.Minute)

	curInterval, curNextPingAt, active, err := s.timerRepo.GetSettings(ctx, userID)
	if err != nil {
		return fmt.Errorf("activate timer: %w", err)
	}
	if active && curInterval == intervalMin && curNextPingAt.After(now) {
		nextPingAt = curNextPingAt
	}

	return s.timerRepo.UpsertInterval(ctx, userID, intervalMin, nextPingAt)
}

// Stop disables timer for user.
func (s *timerService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop timer: invalid userID")
	}
	return s.timerRepo.Disable(ctx, userID)
}

// ListDueUsers returns users that should receive prompt now.
func (s *timerService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.timerRepo.ListDueUsers(ctx, now.UTC(), limit)
}

// MarkPromptSent moves next prompt time forward by interval.
func (s *timerService) MarkPromptSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPingAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return s.timerRepo.SetNextPing(ctx, userID, nextPingAt)
}

// RecordPromptAnswer stores prompt answer using current timer interval from settings.
func (s *timerService) RecordPromptAnswer(ctx context.Context, userID, activityID int64) error {
	intervalMin, err := s.timerRepo.GetInterval(ctx, userID)
	if err != nil {
		return fmt.Errorf("get interval: %w", err)
	}
	return s.sessionRepo.CreateRetroSession(ctx, userID, activityID, intervalMin, "prompt")
}

// RecordPromptAnswerWithInterval stores prompt answer for explicit interval.
func (s *timerService) RecordPromptAnswerWithInterval(ctx context.Context, userID, activityID int64, intervalMin int) error {
	if intervalMin <= 0 {
		return fmt.Errorf("invalid interval")
	}
	return s.sessionRepo.CreateRetroSession(ctx, userID, activityID, intervalMin, "prompt")
}

// AddCustomInterval validates and stores a new custom timer interval.
func (s *timerService) AddCustomInterval(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 {
		return fmt.Errorf("add custom interval: invalid userID")
	}
	if intervalMin < models.MinCustomTimerMinutes || intervalMin > models.MaxCustomTimerMinutes {
		return models.ErrCustomTimerInvalidInterval
	}

	count, err := s.customTimerRepo.Count(ctx, userID)
	if err != nil {
		return fmt.Errorf("add custom interval: %w", err)
	}
	if count >= models.MaxCustomTimersPerUser {
		return models.ErrCustomTimerLimitReached
	}

	return s.customTimerRepo.Create(ctx, userID, intervalMin)
}

// ListCustomIntervals returns the user's custom intervals, ascending.
func (s *timerService) ListCustomIntervals(ctx context.Context, userID int64) ([]int, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("list custom intervals: invalid userID")
	}
	items, err := s.customTimerRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	sort.Ints(items)
	return items, nil
}

// RemoveCustomInterval deletes one previously added custom interval.
func (s *timerService) RemoveCustomInterval(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("remove custom interval: invalid args")
	}
	return s.customTimerRepo.Delete(ctx, userID, intervalMin)
}
