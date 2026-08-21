package learning

import (
	"strings"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

func LearningMenuText(lang i18n.Lang, stats models.LearningStats) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuTitle))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuTotalWords, stats.TotalWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuDueToday, stats.DueTodayWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuLearned, stats.LearnedWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuStreak, stats.StreakDays))
	b.WriteString("\n")
	if stats.TimerActive {
		b.WriteString(i18n.T(lang, i18n.KeyLearningMenuReviewsActive, stats.TimerInterval, stats.NextPushIn))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyLearningMenuReviewsInactive))
	}
	b.WriteString("\n")
	return b.String()
}

// Toggling here while reviews are already running applies immediately, no restart needed.
func LearningReviewPickTitle(lang i18n.Lang, activeCount int, reviewsActive bool, intervalMin int) string {
	if reviewsActive {
		return i18n.T(lang, i18n.KeyLearningReviewPickManageTitle, intervalMin, activeCount)
	}
	if activeCount == 0 {
		return i18n.T(lang, i18n.KeyLearningReviewPickEmptyTitle)
	}
	return i18n.T(lang, i18n.KeyLearningReviewPickTitle, activeCount)
}

func LearningStatsDetailText(lang i18n.Lang, d models.LearningStatsDetail) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyLearningStatsTitle))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuTotalWords, d.Overall.TotalWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuDueToday, d.Overall.DueTodayWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuLearned, d.Overall.LearnedWords))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyLearningMenuStreak, d.Overall.StreakDays))
	b.WriteString("\n")
	if d.ReviewsTotal > 0 {
		accuracy := float64(d.ReviewsCorrect) / float64(d.ReviewsTotal) * 100
		b.WriteString(i18n.T(lang, i18n.KeyLearningStatsAccuracy, accuracy, d.ReviewsCorrect, d.ReviewsTotal))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyLearningStatsNoReviews))
	}
	b.WriteString("\n")

	if len(d.Collections) > 0 {
		b.WriteString("\n")
		b.WriteString(i18n.T(lang, i18n.KeyLearningStatsByCollection))
		b.WriteString("\n")
		for _, c := range d.Collections {
			b.WriteString(i18n.T(lang, i18n.KeyLearningStatsCollectionLine, c.Name, c.TotalWords, c.DueWords, c.LearnedWords))
		}
	}
	return b.String()
}

func LearningRenamePromptText(lang i18n.Lang, currentName string) string {
	return i18n.T(lang, i18n.KeyLearningRenamePrompt, currentName)
}

func LearningWordBaseTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyLearningWordBaseTitle, count)
}

func LearningCollectionDetailTitle(lang i18n.Lang, name string, wordCount int) string {
	return i18n.T(lang, i18n.KeyLearningCollectionDetail, name, wordCount)
}

func LearningArchiveTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyLearningArchiveTitle, count)
}

func LearningReviewCardText(lang i18n.Lang, collectionName, term string) string {
	return i18n.T(lang, i18n.KeyLearningReviewCardTitle, collectionName, term)
}

func LearningReviewRevealedText(lang i18n.Lang, collectionName, term, translation string) string {
	return i18n.T(lang, i18n.KeyLearningReviewRevealed, collectionName, term, translation)
}

// learned overrides the grade message; minute vs day wording is picked from the
// actual delay to nextReviewAt, not the grade (see service.reviewDelay).
func LearningReviewGradedText(lang i18n.Lang, term string, grade models.LearningGrade, nextReviewAt time.Time, learned bool) string {
	if learned {
		return i18n.T(lang, i18n.KeyLearningReviewLearned, term)
	}

	until := time.Until(nextReviewAt)
	useMinutes := until < 24*time.Hour
	minutes := int(until.Minutes() + 0.5)
	if minutes < 1 {
		minutes = 1
	}
	days := int(until.Hours()/24 + 0.5)
	if days < 1 {
		days = 1
	}

	switch grade {
	case models.LearningGradeAgain:
		if useMinutes {
			return i18n.T(lang, i18n.KeyLearningReviewMissedMinutes, term, minutes)
		}
		return i18n.T(lang, i18n.KeyLearningReviewMissed, term)
	case models.LearningGradeHard:
		if useMinutes {
			return i18n.T(lang, i18n.KeyLearningReviewHardConfirmMinutes, term, minutes)
		}
		return i18n.T(lang, i18n.KeyLearningReviewHardConfirm, term, days)
	case models.LearningGradeEasy:
		return i18n.T(lang, i18n.KeyLearningReviewEasyConfirm, term, days)
	default: // models.LearningGradeGood
		return i18n.T(lang, i18n.KeyLearningReviewCorrect, term, days)
	}
}
