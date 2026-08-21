package scheduler

import (
	"context"
	"time"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/service"

	"github.com/rs/zerolog/log"
)

type LearningScheduler struct {
	ctx         context.Context
	learningsvc service.LearningService
	learning    *handlers.Module
}

func NewLearningScheduler(ctx context.Context, learningsvc service.LearningService, learning *handlers.Module) *LearningScheduler {
	return &LearningScheduler{
		ctx:         ctx,
		learningsvc: learningsvc,
		learning:    learning,
	}
}

func (s *LearningScheduler) Run() {
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

func (s *LearningScheduler) tick(now time.Time) {
	dueUsers, err := s.learningsvc.ListDueUsers(s.ctx, now, 100)
	if err != nil {
		log.Error().Err(err).Msg("learning scheduler: list due users failed")
		return
	}

	for _, item := range dueUsers {
		if err := s.learning.SendLearningPromptMessage(s.ctx, item.TgUserID, item.DBUserID); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("learning scheduler: send review card failed")
			continue
		}
		if err := s.learningsvc.MarkPushSent(s.ctx, item.DBUserID, item.IntervalMin, now); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("learning scheduler: mark push sent failed")
		}
	}
}
