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
func LearningReviewPickTitle(activeCount int) string {
	if activeCount == 0 {
		return "🎲 *Pick collections for reviews*\n\nTap to include/exclude — none selected yet. Select at least one, then tap Continue."
	}
	return fmt.Sprintf("🎲 *Pick collections for reviews*\n\n%d selected. Tap to include/exclude, then Continue.", activeCount)
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
