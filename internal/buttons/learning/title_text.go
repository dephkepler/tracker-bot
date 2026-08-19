package learning

import (
	"fmt"
	"strings"
	"tracker-bot/internal/models"
)

// LearningMenuText renders the main Learning screen's stats block.
func LearningMenuText(stats models.LearningStats) string {
	var b strings.Builder
	b.WriteString("🧠 *Learning*\n\n")
	fmt.Fprintf(&b, "📊 Total words: *%d*\n", stats.TotalWords)
	fmt.Fprintf(&b, "📘 Due today: *%d*\n", stats.DueTodayWords)
	fmt.Fprintf(&b, "✅ Learned: *%d*\n", stats.LearnedWords)
	fmt.Fprintf(&b, "🔥 Streak: *%d* day(s)\n", stats.StreakDays)
	if stats.TimerActive {
		fmt.Fprintf(&b, "🕐 Reviews: every *%d* min (next in %s)\n", stats.TimerInterval, stats.NextPushIn)
	} else {
		b.WriteString("🕐 Reviews: not active\n")
	}
	return b.String()
}

// LearningReviewPickTitle renders the review collection-picker's header.
// When reviews are already running, toggling here applies immediately —
// no need to stop and restart the schedule to change what's in rotation.
func LearningReviewPickTitle(activeCount int, reviewsActive bool, intervalMin int) string {
	if reviewsActive {
		return fmt.Sprintf("🔧 *Manage reviews*\n\nRunning every *%d* min. %d collection(s) selected — tap to include/exclude, changes apply immediately.", intervalMin, activeCount)
	}
	if activeCount == 0 {
		return "🎲 *Pick collections for reviews*\n\nTap to include/exclude — none selected yet. Select at least one, then tap Continue."
	}
	return fmt.Sprintf("🎲 *Pick collections for reviews*\n\n%d selected. Tap to include/exclude, then Continue.", activeCount)
}

// LearningStatsDetailText renders the full "📈 Statistics" breakdown: overall
// numbers, per-collection counts, and answer accuracy.
func LearningStatsDetailText(d models.LearningStatsDetail) string {
	var b strings.Builder
	b.WriteString("📈 *Statistics*\n\n")
	fmt.Fprintf(&b, "📊 Total words: *%d*\n", d.Overall.TotalWords)
	fmt.Fprintf(&b, "📘 Due today: *%d*\n", d.Overall.DueTodayWords)
	fmt.Fprintf(&b, "✅ Learned: *%d*\n", d.Overall.LearnedWords)
	fmt.Fprintf(&b, "🔥 Streak: *%d* day(s)\n", d.Overall.StreakDays)
	if d.ReviewsTotal > 0 {
		accuracy := float64(d.ReviewsCorrect) / float64(d.ReviewsTotal) * 100
		fmt.Fprintf(&b, "🎯 Accuracy: *%.0f%%* (%d/%d reviews)\n", accuracy, d.ReviewsCorrect, d.ReviewsTotal)
	} else {
		b.WriteString("🎯 Accuracy: no reviews answered yet\n")
	}

	if len(d.Collections) > 0 {
		b.WriteString("\n*By collection:*\n")
		for _, c := range d.Collections {
			fmt.Fprintf(&b, "• %s — %d words, %d due, %d learned\n", c.Name, c.TotalWords, c.DueWords, c.LearnedWords)
		}
	}
	return b.String()
}

// LearningRenamePromptText asks the user to type a new name for a collection.
func LearningRenamePromptText(currentName string) string {
	return fmt.Sprintf("✏️ Send a new name for *%s* (2-60 characters, single line):", currentName)
}

// LearningWordBaseTitle renders the word-base screen's header.
func LearningWordBaseTitle(count int) string {
	return fmt.Sprintf("🗂 *Word base* — %d collection(s)\n\nTap a collection to view its words.", count)
}

// LearningCollectionDetailTitle renders one collection's detail header.
func LearningCollectionDetailTitle(name string, wordCount int) string {
	return fmt.Sprintf("📚 *%s* — %d word(s)", name, wordCount)
}

// LearningArchiveTitle renders the archive screen's header.
func LearningArchiveTitle(count int) string {
	return fmt.Sprintf("🔁 *Archived collections* — %d", count)
}

// LearningReviewCardText renders the term-only side of a review card.
func LearningReviewCardText(collectionName, term string) string {
	return fmt.Sprintf("🧠 *%s*\n\n%s", collectionName, term)
}

// LearningReviewRevealedText renders the revealed (term + translation) side
// of a review card.
func LearningReviewRevealedText(collectionName, term, translation string) string {
	return fmt.Sprintf("🧠 *%s*\n\n%s\n→ *%s*", collectionName, term, translation)
}

// LearningReviewGradedText renders the confirmation after grading.
func LearningReviewGradedText(term string, correct bool, nextIntervalDays int, learned bool) string {
	if learned {
		return fmt.Sprintf("🎉 *%s* — learned! No more reviews needed.", term)
	}
	if correct {
		return fmt.Sprintf("✅ Nice! Next review of *%s* in %d day(s).", term, nextIntervalDays)
	}
	return fmt.Sprintf("🔁 No worries — *%s* is back in the rotation, next review tomorrow.", term)
}
