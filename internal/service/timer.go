package service

import (
	"context"
	"fmt"
	"sort"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
)

type TimerService interface {
	Activate(ctx context.Context, userID int64, intervalMin int) error
	Stop(ctx context.Context, userID int64) error
	// GetSettings surfaces the raw persisted schedule — enabled=false,
	// err=nil means the user has never activated a timer.
	GetSettings(ctx context.Context, userID int64) (intervalMin int, nextPingAt time.Time, enabled bool, err error)
	ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error)
	MarkPromptSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error
	RecordPromptAnswer(ctx context.Context, userID, activityID int64) error
	// credits the session to endAt (the prompt's original due time), not now — so a late answer still counts for the right window
	RecordPromptAnswerWithInterval(ctx context.Context, userID, activityID int64, intervalMin int, endAt time.Time) error

	// adding an interval that already exists is a no-op (enforced in the repo)
	AddCustomInterval(ctx context.Context, userID int64, intervalMin int) error
	ListCustomIntervals(ctx context.Context, userID int64) ([]int, error)
	RemoveCustomInterval(ctx context.Context, userID int64, intervalMin int) error
}

type timerService struct {
	timerRepo       repo.TimerRepository
	sessionRepo     repo.SessionRepository
	customTimerRepo repo.CustomTimerRepository
}

func NewTimerService(timerRepo repo.TimerRepository, sessionRepo repo.SessionRepository, customTimerRepo repo.CustomTimerRepository) TimerService {
	return &timerService{
		timerRepo:       timerRepo,
		sessionRepo:     sessionRepo,
		customTimerRepo: customTimerRepo,
	}
}

// if already running with the same interval, keeps the existing schedule instead of restarting it
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

func (s *timerService) GetSettings(ctx context.Context, userID int64) (int, time.Time, bool, error) {
	return s.timerRepo.GetSettings(ctx, userID)
}

func (s *timerService) Stop(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("stop timer: invalid userID")
	}
	return s.timerRepo.Disable(ctx, userID)
}

func (s *timerService) ListDueUsers(ctx context.Context, now time.Time, limit int) ([]models.TimerDueUser, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.timerRepo.ListDueUsers(ctx, now.UTC(), limit)
}

func (s *timerService) MarkPromptSent(ctx context.Context, userID int64, intervalMin int, now time.Time) error {
	nextPingAt := now.UTC().Add(time.Duration(intervalMin) * time.Minute)
	return s.timerRepo.SetNextPing(ctx, userID, nextPingAt)
}

// unused by the live prompt flow (see RecordPromptAnswerWithInterval); kept for interface completeness
func (s *timerService) RecordPromptAnswer(ctx context.Context, userID, activityID int64) error {
	intervalMin, err := s.timerRepo.GetInterval(ctx, userID)
	if err != nil {
		return fmt.Errorf("get interval: %w", err)
	}
	return s.sessionRepo.CreateRetroSession(ctx, userID, activityID, intervalMin, "prompt", time.Now().UTC())
}

func (s *timerService) RecordPromptAnswerWithInterval(ctx context.Context, userID, activityID int64, intervalMin int, endAt time.Time) error {
	if intervalMin <= 0 {
		return fmt.Errorf("invalid interval")
	}
	if endAt.IsZero() {
		endAt = time.Now().UTC()
	}
	return s.sessionRepo.CreateRetroSession(ctx, userID, activityID, intervalMin, "prompt", endAt)
}

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

func (s *timerService) RemoveCustomInterval(ctx context.Context, userID int64, intervalMin int) error {
	if userID <= 0 || intervalMin <= 0 {
		return fmt.Errorf("remove custom interval: invalid args")
	}
	return s.customTimerRepo.Delete(ctx, userID, intervalMin)
}
