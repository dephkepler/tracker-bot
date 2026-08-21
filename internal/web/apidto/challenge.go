package apidto

import "tracker-bot/internal/models"

// ChallengeDay is one square of the grid. Status is pending, done or skipped.
type ChallengeDay struct {
	Date   Date   `json:"date"`
	Status string `json:"status"`
}

type Challenge struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Bare dates: challenge_days.day_date is a plain DATE with no zone, so a
	// challenge day is whatever calendar day the bot decided at write time.
	// Sending it as an instant would invite a client to shift it.
	StartDate Date `json:"start_date"`
	EndDate   Date `json:"end_date"`

	TotalDays   int `json:"total_days"`
	DoneDays    int `json:"done_days"`
	SkippedDays int `json:"skipped_days"`
	// PendingDays is derived here so two clients cannot disagree about it.
	PendingDays int `json:"pending_days"`

	CurrentStreak int `json:"current_streak"`
	BestStreak    int `json:"best_streak"`

	Days []ChallengeDay `json:"days"`
}

type ChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
	Meta       Meta        `json:"meta"`
}

type ChallengeDayStateResponse struct {
	ChallengeID int64  `json:"challenge_id"`
	Date        Date   `json:"date"`
	Status      string `json:"status"`
}

func FromChallenge(item models.ChallengeItem, days []models.ChallengeDay, current, best int) Challenge {
	out := Challenge{
		ID:            item.ID,
		Name:          item.Name,
		StartDate:     DateFromNaive(item.StartDate),
		EndDate:       DateFromNaive(item.EndDate),
		TotalDays:     item.TotalDays,
		DoneDays:      item.DoneDays,
		SkippedDays:   item.SkippedDays,
		PendingDays:   max(item.TotalDays-item.DoneDays-item.SkippedDays, 0),
		CurrentStreak: current,
		BestStreak:    best,
		Days:          make([]ChallengeDay, 0, len(days)),
	}
	for _, d := range days {
		out.Days = append(out.Days, ChallengeDay{
			Date:   DateFromNaive(d.Date),
			Status: string(d.Status),
		})
	}
	return out
}
