package apidto

import "tracker-bot/internal/models"

// RoadmapCard is one checklist item.
type RoadmapCard struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
	// Kind is topic, article, book or lecture; Difficulty is 1 easy to 3 hard.
	Kind       string `json:"kind"`
	Difficulty int    `json:"difficulty"`
	Done       bool   `json:"done"`
	// DoneAt is null while pending. It is the only completion timestamp in this
	// domain, and nothing in the bot has ever plotted it — sending it lets the
	// dashboard draw a history the user has never seen.
	DoneAt *Instant `json:"done_at"`
}

// RoadmapTechnology is one technology under a goal.
type RoadmapTechnology struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// MasteryCriteria is the user's own words for what knowing this means.
	MasteryCriteria string `json:"mastery_criteria"`
	// Active is whether it feeds the reminder digest.
	Active     bool          `json:"active"`
	TotalCards int           `json:"total_cards"`
	DoneCards  int           `json:"done_cards"`
	Cards      []RoadmapCard `json:"cards"`
}

// RoadmapGoal is an outcome several technologies feed into.
type RoadmapGoal struct {
	ID           int64               `json:"id"`
	Name         string              `json:"name"`
	TotalCards   int                 `json:"total_cards"`
	DoneCards    int                 `json:"done_cards"`
	Technologies []RoadmapTechnology `json:"technologies"`
}

type RoadmapResponse struct {
	Goals []RoadmapGoal `json:"goals"`
	// Technologies attached to no goal — v1 leftovers, or ones whose goal was
	// deleted. Still real work, so they are not hidden.
	OrphanTechnologies []RoadmapTechnology `json:"orphan_technologies"`
	Meta               Meta                `json:"meta"`
}

type CardStateResponse struct {
	CardID    int64 `json:"card_id"`
	RoadmapID int64 `json:"roadmap_id"`
	Done      bool  `json:"done"`
}

// PlanResult reports what a generation run stored. Rejected counts drafts the
// model produced that failed validation — surfaced rather than hidden, because
// "added 6" when it proposed 12 is worth knowing.
type PlanResult struct {
	Added    int `json:"added"`
	Rejected int `json:"rejected"`
}

type QuizResult struct {
	CardID   int64  `json:"card_id"`
	CardText string `json:"card_text"`
	Question string `json:"question"`
}

type GradeResult struct {
	// Verdict is correct, partial or wrong.
	Verdict  string `json:"verdict"`
	Feedback string `json:"feedback"`
}

// JobResponse is a background AI call. Result is filled once status is "done",
// error_code and error_message once it is "failed".
type JobResponse struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
	Result any    `json:"result,omitempty"`

	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func FromRoadmapCards(cards []models.RoadmapCardItem) []RoadmapCard {
	out := make([]RoadmapCard, 0, len(cards))
	for _, c := range cards {
		card := RoadmapCard{
			ID:         c.ID,
			Text:       c.Text,
			Kind:       string(c.Kind),
			Difficulty: c.Difficulty,
			Done:       c.IsDone,
		}
		if c.DoneAt != nil {
			// A real instant, unlike the bucket columns elsewhere: this one is
			// a genuine timestamptz.
			at := FromInstant(*c.DoneAt)
			card.DoneAt = &at
		}
		out = append(out, card)
	}
	return out
}

func FromRoadmapTechnology(item models.RoadmapItem, cards []models.RoadmapCardItem) RoadmapTechnology {
	return RoadmapTechnology{
		ID:              item.ID,
		Name:            item.Name,
		MasteryCriteria: item.MasteryCriteria,
		Active:          item.Active,
		TotalCards:      item.TotalCards,
		DoneCards:       item.DoneCards,
		Cards:           FromRoadmapCards(cards),
	}
}

func FromRoadmapGoal(goal models.RoadmapGoalItem, techs []RoadmapTechnology) RoadmapGoal {
	return RoadmapGoal{
		ID:           goal.ID,
		Name:         goal.Name,
		TotalCards:   goal.TotalCards,
		DoneCards:    goal.DoneCards,
		Technologies: techs,
	}
}
