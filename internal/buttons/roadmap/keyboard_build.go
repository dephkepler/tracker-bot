package roadmap

import (
	"fmt"
	"strconv"
	"strings"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// cardButtonMaxRunes caps how much of a card's text goes on its button.
// Cards are freeform lines up to models.MaxRoadmapCardTextLen, but Telegram
// renders a long button label as one unreadable smear — the full text stays
// in the DB, the button just shows the start of it.
const cardButtonMaxRunes = 40

// Inline button menus

// RoadmapEntryInlineMenu renders the main Roadmap screen's actions. The
// reminder button's label hints whether the digest push is already running;
// it opens the same interval picker either way.
func RoadmapEntryInlineMenu(lang i18n.Lang, remindersActive bool) tgbotapi.InlineKeyboardMarkup {
	reminderLabel := i18n.T(lang, i18n.KeyRoadmapButtonStartReminders)
	if remindersActive {
		reminderLabel = i18n.T(lang, i18n.KeyRoadmapButtonManageReminders)
	}
	rows := [][]tgbotapi.InlineKeyboardButton{
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonCreate), RoadmapCBCreate),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonList), RoadmapCBList),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(reminderLabel, RoadmapCBPushOpen),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonArchive), RoadmapCBArchiveOpen),
		),
	}
	if remindersActive {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonStopReminders), RoadmapCBPushStop),
		))
	}
	rows = append(rows,
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonProgress), RoadmapCBStats),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home"),
		),
	)
	return buttonbuilder.IK(rows...)
}

// RoadmapBackToMainInlineMenu is a single "back to the Roadmap menu" row,
// used by read-only screens (Progress) that have no other action.
func RoadmapBackToMainInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)),
	)
}

// RoadmapListInlineMenu lists active roadmaps with their done/total counts;
// tapping one opens its card checklist. The 🟢/⚪ marker is the same
// digest-participation flag the detail screen toggles.
func RoadmapListInlineMenu(lang i18n.Lang, items []models.RoadmapItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(RoadmapItemLabel(lang, item), fmt.Sprintf("%s%d", RoadmapCBOpen, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// RoadmapItemLabel renders one roadmap's list-row label, e.g.
// "🟢 Kafka — 3/12".
func RoadmapItemLabel(lang i18n.Lang, item models.RoadmapItem) string {
	key := i18n.KeyRoadmapButtonToggleOffFmt
	if item.Active {
		key = i18n.KeyRoadmapButtonToggleOnFmt
	}
	return i18n.T(lang, key, item.Name, item.DoneCards, item.TotalCards)
}

// RoadmapDetailInlineMenu shows one roadmap's cards — each row is the card
// itself (tap to tick/untick) plus a delete button — followed by the
// roadmap-level actions.
func RoadmapDetailInlineMenu(lang i18n.Lang, roadmapID int64, active bool, cards []models.RoadmapCardItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cards)+5)
	for _, c := range cards {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(CardButtonLabel(c), fmt.Sprintf("%s%d", RoadmapCBCardToggle, c.ID)),
			buttonbuilder.IB("🗑", fmt.Sprintf("%s%d", RoadmapCBCardDelete, c.ID)),
		))
	}

	toggleLabel := i18n.T(lang, i18n.KeyRoadmapLabelExcludedReminder)
	if active {
		toggleLabel = i18n.T(lang, i18n.KeyRoadmapLabelIncludedInReminder)
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(toggleLabel, fmt.Sprintf("%s%d", RoadmapCBToggle, roadmapID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonAddCards), fmt.Sprintf("%s%d", RoadmapCBAddCards, roadmapID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonSetGoal), fmt.Sprintf("%s%d", RoadmapCBSetGoal, roadmapID)),
	))
	rows = append(rows, buttonbuilder.IR(
		// "✏️ Rename" reuses the Learning key on purpose — identical text,
		// and two keys rendering the same string would collide in i18n's
		// reverse index (see TestCatalog_NoTextCollisionsWithinLanguage).
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonRename), fmt.Sprintf("%s%d", RoadmapCBRename, roadmapID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonArchiveThis), fmt.Sprintf("%s%d", RoadmapCBArchiveThis, roadmapID)),
	))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBList)))
	return buttonbuilder.IK(rows...)
}

// CardButtonLabel renders one card's button text: a done/pending box plus
// the (possibly shortened) card text.
func CardButtonLabel(c models.RoadmapCardItem) string {
	box := "⬜"
	if c.IsDone {
		box = "✅"
	}
	return box + " " + truncateRunes(c.Text, cardButtonMaxRunes)
}

// truncateRunes shortens s to at most maxRunes runes, appending an ellipsis
// when it had to cut. Rune-based, not byte-based, so a Cyrillic or Arabic
// card isn't cut mid-character.
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}

// RoadmapArchiveInlineMenu lists archived roadmaps with restore/delete
// actions — mirrors learning.LearningArchiveInlineMenu.
func RoadmapArchiveInlineMenu(lang i18n.Lang, items []models.RoadmapItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+2)
	for _, item := range items {
		label := i18n.T(lang, i18n.KeyRoadmapArchiveItemFmt, item.Name, item.DoneCards, item.TotalCards)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", RoadmapCBArchiveRestore, item.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", RoadmapCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), RoadmapCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// RoadmapDigestInlineMenu gives every card in a reminder push its own tick
// button, so a card can be closed straight from the notification without
// navigating into the bot's menus. It uses RoadmapCBDigestToggle rather
// than the card list's own prefix so the handler knows to re-render this
// digest message instead of a checklist screen.
func RoadmapDigestInlineMenu(lang i18n.Lang, cards []models.RoadmapDigestCard) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(cards)+1)
	for _, c := range cards {
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB("✅ "+truncateRunes(c.Text, cardButtonMaxRunes), fmt.Sprintf("%s%d", RoadmapCBDigestToggle, c.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyRoadmapButtonList), RoadmapCBList),
	))
	return buttonbuilder.IK(rows...)
}

// Reply button menus

// RoadmapWaitingReplyMenu is shown while the bot waits for typed input (a
// roadmap name) — lets the user bail out.
func RoadmapWaitingReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

// RoadmapGoalReplyMenu is shown while waiting for a mastery goal — the goal
// is optional, so Skip sits next to Cancel.
func RoadmapGoalReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyRoadmapButtonSkipGoal))),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

// RoadmapAddCardsReplyMenu is shown while pasting card lines — "Done"
// leaves the flow, since the user may paste several messages in a row.
func RoadmapAddCardsReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonDone))),
	)
}

// pushIntervalPrefix marks a reminder-interval reply button, e.g.
// "⏰ 180 min". Deliberately a different emoji than Learning's "⏱ " so the
// two pickers' buttons never resolve to each other if one is left on
// screen — dispatch is screen-gated either way, but this keeps it obvious.
const pushIntervalPrefix = "⏰ "

// RoadmapPushIntervalReplyMenu renders the reminder-interval picker.
func RoadmapPushIntervalReplyMenu(lang i18n.Lang, intervals []int) tgbotapi.ReplyKeyboardMarkup {
	rows := make([][]tgbotapi.KeyboardButton, 0, (len(intervals)+1)/2+1)
	for i := 0; i < len(intervals); i += 2 {
		if i+1 < len(intervals) {
			rows = append(rows, buttonbuilder.RR(
				buttonbuilder.RB(FormatPushIntervalButton(intervals[i])),
				buttonbuilder.RB(FormatPushIntervalButton(intervals[i+1])),
			))
		} else {
			rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(FormatPushIntervalButton(intervals[i]))))
		}
	}
	rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonBack))))
	return buttonbuilder.RK(rows...)
}

// FormatPushIntervalButton renders one interval as reply-button text, e.g.
// "⏰ 180 min". Not translated — a plain unit abbreviation, same reasoning
// as learning.FormatPushIntervalButton's own "min" suffix.
func FormatPushIntervalButton(minutes int) string {
	return fmt.Sprintf("%s%d min", pushIntervalPrefix, minutes)
}

// ParsePushIntervalButtonMinutes is the inverse of FormatPushIntervalButton.
func ParsePushIntervalButtonMinutes(text string) (int, bool) {
	if !strings.HasPrefix(text, pushIntervalPrefix) {
		return 0, false
	}
	if !strings.HasSuffix(text, " min") {
		return 0, false
	}
	numStr := strings.TrimSuffix(strings.TrimPrefix(text, pushIntervalPrefix), " min")
	n, err := strconv.Atoi(numStr)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
