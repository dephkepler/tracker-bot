package scheduler

import (
	"context"
	"time"
	"tracker-bot/internal/handlers"
	"tracker-bot/internal/service"

	"github.com/rs/zerolog/log"
)

// ChallengeScheduler periodically checks due challenges and sends the daily
// evening "did you do it?" push — mirrors TimerScheduler/LearningScheduler.
type ChallengeScheduler struct {
	ctx          context.Context
	challengesvc service.ChallengeService
	challenge    *handlers.Module
}

// NewChallengeScheduler creates scheduler instance.
func NewChallengeScheduler(ctx context.Context, challengesvc service.ChallengeService, challenge *handlers.Module) *ChallengeScheduler {
	return &ChallengeScheduler{
		ctx:          ctx,
		challengesvc: challengesvc,
		challenge:    challenge,
	}
}

// Run starts background ticker loop. A once-a-day push doesn't need tight
// polling — 60s keeps drift small without hammering the DB.
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

// tick processes one scheduler cycle at provided UTC time.
func (s *ChallengeScheduler) tick(now time.Time) {
	due, err := s.challengesvc.ListDueChallenges(s.ctx, now, 100)
	if err != nil {
		log.Error().Err(err).Msg("challenge scheduler: list due challenges failed")
		return
	}

	for _, item := range due {
		// SendChallengePush also reschedules tomorrow's push (using the
		// user's own timezone) regardless of whether it actually sent —
		// see its doc comment.
		if err := s.challenge.SendChallengePush(s.ctx, item.TgUserID, item.DBUserID, item.ChallengeID, item.ChallengeName, item.StartDate, item.EndDate); err != nil {
			log.Error().Err(err).Int64("challenge_id", item.ChallengeID).Msg("challenge scheduler: send push failed")
		}
	}
}
