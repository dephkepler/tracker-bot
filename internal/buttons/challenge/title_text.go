package challenge

import (
	"fmt"
	"time"
	"tracker-bot/internal/models"
)

// ListTitle renders the challenges list screen's header.
func ListTitle(count int) string {
	if count == 0 {
		return "🎯 *Challenges*\n\nNo challenges yet. Start one — e.g. \"100 days of reading\"."
	}
	return fmt.Sprintf("🎯 *Challenges* — %d active", count)
}

// GridTitle renders one challenge's grid header: name, date range, and
// progress.
func GridTitle(item models.ChallengeItem) string {
	pct := 0
	if item.TotalDays > 0 {
		pct = item.DoneDays * 100 / item.TotalDays
	}
	pending := item.TotalDays - item.DoneDays - item.SkippedDays
	return fmt.Sprintf(
		"🎯 *%s*\n%s → %s (%d days)\n\n✅ %d done · ❌ %d skipped · 🔲 %d pending\n📈 %d%% complete\n\nTap a square to mark that day.",
		item.Name, item.StartDate.Format("2006-01-02"), item.EndDate.Format("2006-01-02"), item.TotalDays,
		item.DoneDays, item.SkippedDays, pending, pct,
	)
}

// DayConfirmTitle renders the "mark this day" screen's header.
func DayConfirmTitle(challengeName string, day time.Time, current models.ChallengeDayStatus) string {
	status := "not marked yet"
	switch current {
	case models.ChallengeDayDone:
		status = "currently marked ✅ Done"
	case models.ChallengeDaySkipped:
		status = "currently marked ❌ Skipped"
	}
	return fmt.Sprintf("🎯 *%s*\n\n%s — %s.\n\nMark this day:", challengeName, day.Format("Mon, 2 Jan 2006"), status)
}

// ArchiveTitle renders the archive screen's header.
func ArchiveTitle(count int) string {
	return fmt.Sprintf("🔁 *Archived challenges* — %d", count)
}

// PushText renders the daily evening push message.
func PushText(challengeName string, dayNum, totalDays int) string {
	return fmt.Sprintf("🎯 *%s* — Day %d/%d\n\nDid you do it today?", challengeName, dayNum, totalDays)
}

// CreatePromptNameText asks for a new challenge's name.
const CreatePromptNameText = "✏️ Name your challenge (e.g. \"100 days of reading\"), 2-60 characters, single line:"

// CreatePromptRangeText asks the user to pick the challenge's date range.
const CreatePromptRangeText = "📅 Pick the start date, then the end date (max 100 days total):"

// CreatedText confirms a challenge was created.
func CreatedText(name string, totalDays int) string {
	return fmt.Sprintf("🎯 Challenge *%s* created — %d days. First check-in tonight at 21:00.", name, totalDays)
}
