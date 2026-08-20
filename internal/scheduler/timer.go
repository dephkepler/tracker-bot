package scheduler

import (
	"context"
	"time"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/service"

	"github.com/rs/zerolog/log"
)

// TimerScheduler periodically checks due users and sends tracking prompts.
type TimerScheduler struct {
	ctx      context.Context
	timersvc service.TimerService
	track    *handlers.Module
}

// NewTimerScheduler creates scheduler instance.
func NewTimerScheduler(ctx context.Context, timersvc service.TimerService, track *handlers.Module) *TimerScheduler {
	return &TimerScheduler{
		ctx:      ctx,
		timersvc: timersvc,
		track:    track,
	}
}

// Run starts background ticker loop.
func (s *TimerScheduler) Run() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case now := <-ticker.C:
				s.tick(now.UTC())
			}
		}
	}()
}

// tick processes one scheduler cycle at provided UTC time.
func (s *TimerScheduler) tick(now time.Time) {
	dueUsers, err := s.timersvc.ListDueUsers(s.ctx, now, 100)
	if err != nil {
		log.Error().Err(err).Msg("timer scheduler: list due users failed")
		return
	}

	for _, item := range dueUsers {
		if err := s.track.SendPromptMessage(s.ctx, item.TgUserID, item.DBUserID, item.IntervalMin, item.NextPingAt); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("timer scheduler: send prompt failed")
			continue
		}
		if err := s.timersvc.MarkPromptSent(s.ctx, item.DBUserID, item.IntervalMin, now); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("timer scheduler: mark prompt sent failed")
		}
	}
}
