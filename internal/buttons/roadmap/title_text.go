package roadmap

import (
	"strings"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
)

func RoadmapMenuText(lang i18n.Lang, stats models.RoadmapStats) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTitle))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalGoals, stats.TotalGoals, models.MaxRoadmapGoalsPerUser))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalRoadmaps, stats.TotalRoadmaps))
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

func RoadmapGoalsTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyRoadmapGoalsTitle, count, models.MaxRoadmapGoalsPerUser)
}

// RoadmapGoalDetailText renders a goal's header: its name, how many
// technologies and cards it holds, and how far along it is.
func RoadmapGoalDetailText(lang i18n.Lang, goal models.RoadmapGoalItem) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapGoalDetailTitle, goal.Name))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapGoalDetailCounts,
		goal.TotalRoadmaps, models.MaxRoadmapsPerGoal,
		goal.DoneCards, goal.TotalCards, PercentDone(goal.DoneCards, goal.TotalCards)))
	b.WriteString("\n\n")
	if goal.TotalRoadmaps == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapGoalDetailNoTech))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapGoalDetailHint))
	}
	return b.String()
}

func RoadmapListTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyRoadmapListTitle, count, models.MaxRoadmapsPerGoal)
}

func RoadmapOrphansTitle(lang i18n.Lang, count int) string {
	return i18n.T(lang, i18n.KeyRoadmapOrphansTitle, count)
}

// RoadmapDetailText renders one technology's header: name, done/total, its
// mastery criteria, and a hint about the card row's buttons.
// cardTextMaxRunes caps one card in the list — generous, because a card is
// meant to be read here.
const cardTextMaxRunes = 160

// cardListBudgetRunes bounds the list as a whole, which is the constraint that
// actually matters: a Telegram message holds 4096 characters, and capping only
// the per-card length let thirty long cards add up to five thousand and the
// whole screen fail to send. Budgeting the total adapts to however long the
// cards happen to be instead of relying on a product of two guesses.
const cardListBudgetRunes = 3000

func RoadmapDetailText(lang i18n.Lang, item models.RoadmapItem, cards []models.RoadmapCardItem) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailTitle, item.Name, item.DoneCards, item.TotalCards))
	b.WriteString("\n")
	if strings.TrimSpace(item.MasteryCriteria) == "" {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailNoCriteria))
	} else {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailCriteria, item.MasteryCriteria))
	}
	b.WriteString("\n\n")
	if len(cards) == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailNoCards))
		return b.String()
	}

	// The cards are listed here rather than left to their buttons: an inline
	// button is roughly 35 characters wide, and a generated card is a sentence.
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailCardsHdr))
	b.WriteString("\n")
	shown, spent := 0, 0
	for i, card := range cards {
		line := i18n.T(lang, i18n.KeyRoadmapDetailCardLine,
			i+1, CardStateIcons(card), truncateRunes(card.Text, cardTextMaxRunes))
		// At least one card is always listed, however long it is: an empty list
		// under a "Cards" header would be stranger than a truncated one.
		if shown > 0 && spent+len([]rune(line)) > cardListBudgetRunes {
			break
		}
		b.WriteString(line)
		b.WriteString("\n")
		spent += len([]rune(line)) + 1
		shown++
	}
	if len(cards) > shown {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailCardsMore, len(cards)-shown))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapDetailHint))
	return b.String()
}

func RoadmapCriteriaPromptText(lang i18n.Lang, name string) string {
	return i18n.T(lang, i18n.KeyRoadmapCriteriaPrompt, name, models.MaxRoadmapCriteriaLen)
}

func RoadmapRenamePromptText(lang i18n.Lang, currentName string) string {
	return i18n.T(lang, i18n.KeyRoadmapRenamePrompt, currentName)
}

func RoadmapGoalRenamePromptText(lang i18n.Lang, currentName string) string {
	return i18n.T(lang, i18n.KeyRoadmapGoalRenamePrompt, currentName)
}

func RoadmapAssignPromptText(lang i18n.Lang, name string) string {
	return i18n.T(lang, i18n.KeyRoadmapAssignPrompt, name)
}

// RoadmapArchiveText renders the archive header plus a section label for
// whichever of goals/technologies is actually present.
func RoadmapArchiveText(lang i18n.Lang, goals []models.RoadmapGoalItem, items []models.RoadmapItem) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapArchiveTitle))
	b.WriteString("\n")
	if len(goals) > 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapArchiveGoalsHdr))
		b.WriteString("\n")
	}
	if len(items) > 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapArchiveTechHdr))
		b.WriteString("\n")
	}
	return b.String()
}

// RoadmapDigestText renders a reminder push: the pending cards grouped by
// technology, in the easiest-first order PickDigestCards returned them.
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
		b.WriteString(difficultyIcon(c.Difficulty))
		b.WriteString(kindIcon(c.Kind))
		b.WriteString(" ")
		b.WriteString(c.Text)
		b.WriteString("\n")
	}
	return b.String()
}

// RoadmapStatsDetailText renders the progress breakdown: overall numbers,
// then each goal with the technologies underneath it. Grouping is done off
// the per-technology rows' GoalName, which the repo already returns ordered
// by goal.
func RoadmapStatsDetailText(lang i18n.Lang, d models.RoadmapStatsDetail) string {
	var b strings.Builder
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsTitle))
	b.WriteString("\n\n")

	if len(d.Roadmaps) == 0 && len(d.Goals) == 0 {
		b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsEmpty))
		return b.String()
	}

	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalGoals, d.Overall.TotalGoals, models.MaxRoadmapGoalsPerUser))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuTotalCards, d.Overall.TotalCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuDone, d.Overall.DoneCards))
	b.WriteString("\n")
	b.WriteString(i18n.T(lang, i18n.KeyRoadmapMenuPending, d.Overall.PendingCards))
	b.WriteString("\n")

	goalTotals := make(map[string][2]int)
	order := make([]string, 0)
	for _, r := range d.Roadmaps {
		if _, seen := goalTotals[r.GoalName]; !seen {
			order = append(order, r.GoalName)
		}
		t := goalTotals[r.GoalName]
		goalTotals[r.GoalName] = [2]int{t[0] + r.DoneCards, t[1] + r.TotalCards}
	}

	for _, goalName := range order {
		totals := goalTotals[goalName]
		if goalName == "" {
			b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsNoGoalHeader))
		} else {
			b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsGoalLine, goalName, totals[0], totals[1], PercentDone(totals[0], totals[1])))
		}
		for _, r := range d.Roadmaps {
			if r.GoalName != goalName {
				continue
			}
			b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsRoadmapLine, r.Name, r.DoneCards, r.TotalCards, PercentDone(r.DoneCards, r.TotalCards)))
			if strings.TrimSpace(r.MasteryCriteria) != "" {
				b.WriteString(i18n.T(lang, i18n.KeyRoadmapStatsCriteriaLine, r.MasteryCriteria))
			}
		}
	}
	return b.String()
}
