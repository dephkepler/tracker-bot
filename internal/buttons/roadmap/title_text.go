package roadmap

import (
	"strings"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

// RoadmapMenuText renders the main Roadmap screen's stats block.
func RoadmapMenuText(lang i18n.Lang, stats models.RoadmapStats) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTitle))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalRoadmaps, stats.TotalRoadmaps, models.MaxRoadmapsPerUser))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalCards, stats.TotalCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuDone, stats.DoneCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuPending, stats.PendingCards))
	b.WriteString("\n")
	if stats.TimerActive {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuRemindersActive, stats.TimerInterval, stats.NextPushIn))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuRemindersInactive))
	}
	b.WriteString("\n")
	return b.String()
}

// RoadmapListTitle renders the roadmap-list screen's header.
func RoadmapListTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyRoadmapListTitle, count, models.MaxRoadmapsPerUser)
}

// RoadmapDetailText renders one roadmap's header: name, done/total, its
// mastery goal, and a hint about ticking cards.
func RoadmapDetailText(lang i18n.Lang, item models.RoadmapItem, cardCount int) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailTitle, item.Name, item.DoneCards, item.TotalCards))
	b.WriteString("\n")
	if strings.TrimSpace(item.Goal) == "" {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailNoGoal))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailGoal, item.Goal))
	}
	b.WriteString("\n\n")
	if cardCount == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailNoCards))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailHint))
	}
	return b.String()
}

// RoadmapArchiveTitle renders the archive screen's header.
func RoadmapArchiveTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyRoadmapArchiveTitle, count)
}

// RoadmapGoalPromptText asks for a roadmap's mastery goal.
func RoadmapGoalPromptText(lang i18n.Lang, name string) string {
	return i18n.T(lang, i18n.KeyRoadmapGoalPrompt, name, models.MaxRoadmapGoalLen)
}

// RoadmapRenamePromptText asks for a new name for a roadmap.
func RoadmapRenamePromptText(lang i18n.Lang, currentName string) string {
	return i18n.T(lang, i18n.KeyRoadmapRenamePrompt, currentName)
}

// RoadmapDigestText renders a reminder push: pending cards grouped by
// roadmap, in the order PickDigestCards returned them. Each card also gets
// its own tick button (see RoadmapDigestInlineMenu) — the text is the
// readable version, the buttons are the actionable one.
func RoadmapDigestText(lang i18n.Lang, cards []models.RoadmapDigestCard, byRoadmap map[int64]models.RoadmapItem) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapDigestTitle))
	b.WriteString("\n")

	lastRoadmapID := int64(0)
	for _, c := range cards {
		if c.RoadmapID != lastRoadmapID {
			item := byRoadmap[c.RoadmapID]
			b.WriteString(i18n.T(lang, i18n.KeyRoadmapDigestRoadmapLine, c.RoadmapName, item.DoneCards, item.TotalCards))
			lastRoadmapID = c.RoadmapID
		}
		b.WriteString("• ")
		b.WriteString(c.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// RoadmapStatsDetailText renders the "📈 Progress" breakdown: overall
// numbers plus a per-technology line with its completion percentage and
// goal.
func RoadmapStatsDetailText(lang i18n.Lang, d models.RoadmapStatsDetail) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsTitle))
	b.WriteString("\n\n")

	if len(d.Roadmaps) == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsEmpty))
		return b.String()
	}

	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalRoadmaps, d.Overall.TotalRoadmaps, models.MaxRoadmapsPerUser))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalCards, d.Overall.TotalCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuDone, d.Overall.DoneCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuPending, d.Overall.PendingCards))
	b.WriteString("\n\n")

	for _, r := range d.Roadmaps {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsRoadmapLine, r.Name, r.DoneCards, r.TotalCards, percentDone(r.DoneCards, r.TotalCards)))
		if strings.TrimSpace(r.Goal) != "" {
			b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsGoalLine, r.Goal))
		}
	}
	return b.String()
}

// percentDone is done/total as a whole percentage, 0 for an empty roadmap
// (rather than a division by zero).
func percentDone(done, total int) int {
	if total <= 0 {
		return 0
	}
	return done * 100 / total
}
