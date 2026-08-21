package apidto

import (
	"math"
	"time"

	"tracker-bot/internal/models"
)

// LearningCollection is one word collection's progress.
type LearningCollection struct {
	Name         string `json:"name"`
	TotalWords   int    `json:"total_words"`
	DueWords     int    `json:"due_words"`
	LearnedWords int    `json:"learned_words"`
}

// DayCount is one day of review activity.
type DayCount struct {
	Date    Date `json:"date"`
	Total   int  `json:"total"`
	Correct int  `json:"correct"`
}

type LearningResponse struct {
	TotalWords   int `json:"total_words"`
	DueWords     int `json:"due_words"`
	LearnedWords int `json:"learned_words"`
	StreakDays   int `json:"streak_days"`

	ReviewsTotal   int `json:"reviews_total"`
	ReviewsCorrect int `json:"reviews_correct"`
	// AccuracyPercent is over all reviews ever, one decimal. Zero reviews give
	// zero rather than a division by nothing.
	AccuracyPercent float64 `json:"accuracy_percent"`

	Collections []LearningCollection `json:"collections"`
	// ReviewsByDay is a continuous run of days, gaps included as zeros: a
	// missing day is the interesting part of a review habit, so it must not be
	// omitted and quietly closed up.
	ReviewsByDay []DayCount `json:"reviews_by_day"`

	// Reminder is the push schedule. Deliberately not the service's own
	// NextPushIn, which is a pre-formatted "%d min" string built for a Telegram
	// message; an API sends the number and lets the client phrase it.
	ReminderActive   bool `json:"reminder_active"`
	ReminderInterval int  `json:"reminder_interval_minutes"`

	Meta Meta `json:"meta"`
}

// FromLearningDetail maps the stats. reviewsByDay is built by the handler,
// which is the only place that knows the requested window.
func FromLearningDetail(detail models.LearningStatsDetail, reviewsByDay []DayCount, meta Meta) LearningResponse {
	collections := make([]LearningCollection, 0, len(detail.Collections))
	for _, c := range detail.Collections {
		collections = append(collections, LearningCollection{
			Name:         c.Name,
			TotalWords:   c.TotalWords,
			DueWords:     c.DueWords,
			LearnedWords: c.LearnedWords,
		})
	}

	accuracy := 0.0
	if detail.ReviewsTotal > 0 {
		accuracy = math.Round(float64(detail.ReviewsCorrect)/float64(detail.ReviewsTotal)*1000) / 10
	}

	return LearningResponse{
		TotalWords:       detail.Overall.TotalWords,
		DueWords:         detail.Overall.DueTodayWords,
		LearnedWords:     detail.Overall.LearnedWords,
		StreakDays:       detail.Overall.StreakDays,
		ReviewsTotal:     detail.ReviewsTotal,
		ReviewsCorrect:   detail.ReviewsCorrect,
		AccuracyPercent:  accuracy,
		Collections:      collections,
		ReviewsByDay:     reviewsByDay,
		ReminderActive:   detail.Overall.TimerActive,
		ReminderInterval: detail.Overall.TimerInterval,
		Meta:             meta,
	}
}

// CountReviewsByDay buckets reviews into the user's own local days, filling
// gaps with zeros across the whole window.
//
// reviewed_at is a genuine instant, unlike the zoneless values the track
// aggregates return, so this converts it into loc — the one place in this
// package where converting is right rather than wrong.
func CountReviewsByDay(entries []models.LearningReviewEntry, from, to time.Time, loc *time.Location) []DayCount {
	type counts struct{ total, correct int }
	byDay := make(map[string]counts, 64)
	for _, e := range entries {
		key := e.ReviewedAt.In(loc).Format("2006-01-02")
		c := byDay[key]
		c.total++
		if e.Correct {
			c.correct++
		}
		byDay[key] = c
	}

	out := make([]DayCount, 0, 32)
	for day := from; day.Before(to); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		c := byDay[key]
		out = append(out, DayCount{Date: Date(key), Total: c.total, Correct: c.correct})
	}
	return out
}
