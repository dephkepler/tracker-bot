package scheduler

import (
	"context"
	"time"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/service"

	"github.com/rs/zerolog/log"
)

type RoadmapScheduler struct {
	ctx        context.Context
	roadmapsvc service.RoadmapService
	roadmap    *handlers.Module
}

func NewRoadmapScheduler(ctx context.Context, roadmapsvc service.RoadmapService, roadmap *handlers.Module) *RoadmapScheduler {
	return &RoadmapScheduler{
		ctx:        ctx,
		roadmapsvc: roadmapsvc,
		roadmap:    roadmap,
	}
}

func (s *RoadmapScheduler) Run() {
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

func (s *RoadmapScheduler) tick(now time.Time) {
	dueUsers, err := s.roadmapsvc.ListDueUsers(s.ctx, now, 100)
	if err != nil {
		log.Error().Err(err).Msg("roadmap scheduler: list due users failed")
		return
	}

	for _, item := range dueUsers {
		// nil here can mean "nothing pending, nothing sent" — schedule still advances so an all-done user isn't retried every tick.
		if err := s.roadmap.SendRoadmapDigestMessage(s.ctx, item.TgUserID, item.DBUserID); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("roadmap scheduler: send digest failed")
			continue
		}
		if err := s.roadmapsvc.MarkPushSent(s.ctx, item.DBUserID, item.IntervalMin, now); err != nil {
			log.Error().Err(err).Int64("user_id", item.DBUserID).Msg("roadmap scheduler: mark push sent failed")
		}
	}
}
