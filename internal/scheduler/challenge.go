package scheduler

import (
	"context"
	"time"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/service"

	"github.com/rs/zerolog/log"
)

type ChallengeScheduler struct {
	ctx          context.Context
	challengesvc service.ChallengeService
	challenge    *handlers.Module
}

func NewChallengeScheduler(ctx context.Context, challengesvc service.ChallengeService, challenge *handlers.Module) *ChallengeScheduler {
	return &ChallengeScheduler{
		ctx:          ctx,
		challengesvc: challengesvc,
		challenge:    challenge,
	}
}

// 60s poll is plenty for a once-a-day push, without hammering the DB.
func (s *ChallengeScheduler) Run() {
	ticker := time.NewTicker(60 * time.Second)
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

func (s *ChallengeScheduler) tick(now time.Time) {
	due, err := s.challengesvc.ListDueChallenges(s.ctx, now, 100)
	if err != nil {
		log.Error().Err(err).Msg("challenge scheduler: list due challenges failed")
		return
	}

	for _, item := range due {
		// SendChallengePush reschedules tomorrow's push regardless of whether this one sent.
		if err := s.challenge.SendChallengePush(s.ctx, item.TgUserID, item.DBUserID, item.ChallengeID, item.ChallengeName, item.StartDate, item.EndDate); err != nil {
			log.Error().Err(err).Int64("challenge_id", item.ChallengeID).Msg("challenge scheduler: send push failed")
		}
	}
}
