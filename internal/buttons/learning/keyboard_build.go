package learning

import (
	"fmt"
	"strconv"
	"strings"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

// LearningEntryInlineMenu renders the main Learning screen's actions. The
// review button always opens the same collection picker (see
// LearningReviewPickInlineMenu) — its label just hints whether reviews are
// already running, so changing what's in rotation never requires stopping
// the schedule first.
func LearningEntryInlineMenu(reviewsActive bool) tgbotapi.InlineKeyboardMarkup {
	reviewLabel := LearningButtonStartReviews
	if reviewsActive {
		reviewLabel = LearningButtonManageReviews
	}
	reviewBtn := buttonbuilder.IB(reviewLabel, LearningCBReviewOpen)
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonCreateCollection, LearningCBCreateCollection),
			buttonbuilder.IB(LearningButtonWordBase, LearningCBWordBase),
		),
		buttonbuilder.IR(
			reviewBtn,
			buttonbuilder.IB(LearningButtonArchive, LearningCBArchiveOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonStatistics, LearningCBStats),
		),
		buttonbuilder.IR(
			buttonbuilder.IB("🏠 Home", "go_home"),
		),
	)
}

// LearningBackToMainInlineMenu is a single "back to the Learning menu" row,
// used by read-only screens (e.g. Statistics) that have no other action.
func LearningBackToMainInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(LearningButtonBack, LearningCBBackMain)),
	)
}

// LearningWordBaseInlineMenu lists active collections; tapping one opens its
// detail view (see LearningCollectionDetailInlineMenu).
func LearningWordBaseInlineMenu(items []models.LearningCollectionItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		label := fmt.Sprintf(LearningButtonToggleOffFmt, item.Name, item.WordCount)
		if item.Active {
			label = fmt.Sprintf(LearningButtonToggleOnFmt, item.Name, item.WordCount)
		}
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, fmt.Sprintf("%s%d", LearningCBCollectionOpen, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonBack, LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// LearningReviewPickInlineMenu lets the user choose which collections feed
// the review rotation — each row toggles the same is_active flag as
// LearningWordBaseInlineMenu, just scoped to this screen so the toggle
// re-renders the picker instead of navigating into a collection's detail
// view. Toggling here takes effect immediately, whether reviews are
// running or not — when active, the bottom row offers Stop instead of
// Continue, since there's no separate "apply" step.
func LearningReviewPickInlineMenu(items []models.LearningCollectionItem, active bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+2)
	for _, item := range items {
		label := fmt.Sprintf(LearningButtonToggleOffFmt, item.Name, item.WordCount)
		if item.Active {
			label = fmt.Sprintf(LearningButtonToggleOnFmt, item.Name, item.WordCount)
		}
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, fmt.Sprintf("%s%d", LearningCBReviewPickToggle, item.ID)),
		))
	}
	if active {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonStopReviews, LearningCBReviewStop)))
	} else {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonContinue, LearningCBReviewContinue)))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonBack, LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// LearningCollectionDetailInlineMenu shows one collection's words (each with
// a delete button) plus collection-level actions.
func LearningCollectionDetailInlineMenu(collectionID int64, active bool, words []models.LearningWordItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(words)+3)
	for _, w := range words {
		label := fmt.Sprintf("%s → %s", w.Term, w.Translation)
		if w.Learned {
			label = "✅ " + label
		}
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, "noop"),
			buttonbuilder.IB("🗑", fmt.Sprintf("%s%d", LearningCBWordDelete, w.ID)),
		))
	}

	toggleLabel := "🟢 Included in reviews"
	if !active {
		toggleLabel = "⚪ Excluded from reviews"
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(toggleLabel, fmt.Sprintf("%s%d", LearningCBCollectionToggle, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(LearningButtonAddWords, fmt.Sprintf("%s%d", LearningCBCollectionAddMore, collectionID)),
		buttonbuilder.IB(LearningButtonRename, fmt.Sprintf("%s%d", LearningCBCollectionRename, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(LearningButtonArchiveThis, fmt.Sprintf("%s%d", LearningCBCollectionArchive, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonBack, LearningCBWordBase)))
	return buttonbuilder.IK(rows...)
}

// LearningArchiveInlineMenu lists archived collections with restore/delete
// actions — mirrors track.TrackArchiveInlineMenu.
func LearningArchiveInlineMenu(items []models.LearningCollectionItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+2)
	for _, item := range items {
		label := fmt.Sprintf("📦 %s — %d words", item.Name, item.WordCount)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonRestore, fmt.Sprintf("%s%d", LearningCBArchiveRestore, item.ID)),
			buttonbuilder.IB(LearningButtonDeleteForever, fmt.Sprintf("%s%d", LearningCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LearningButtonBack, LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// LearningReviewRevealInlineMenu is the first state of a review card: only
// the term is shown, tap to reveal the translation.
func LearningReviewRevealInlineMenu(wordID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(LearningButtonShowAnswer, fmt.Sprintf("%s%d", LearningCBReviewReveal, wordID))),
	)
}

// LearningReviewGradeInlineMenu is the revealed state: user grades whether
// they knew the word.
func LearningReviewGradeInlineMenu(wordID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonKnewIt, fmt.Sprintf("%s%d", LearningCBReviewKnew, wordID)),
			buttonbuilder.IB(LearningButtonMissedIt, fmt.Sprintf("%s%d", LearningCBReviewMissed, wordID)),
		),
	)
}

// Reply button menus

// LearningWaitingReplyMenu is shown while the bot is waiting for typed
// input (collection name, word lines) — lets the user bail out.
func LearningWaitingReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LearningButtonCancel)),
	)
}

// LearningAddWordsReplyMenu is shown while pasting word lines — "Done"
// leaves the flow, unlike single-shot prompts which auto-exit.
func LearningAddWordsReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LearningButtonDone)),
	)
}

// pushIntervalPrefix marks a review-push interval reply button, e.g. "⏱ 60 min".
const pushIntervalPrefix = "⏱ "

// LearningPushIntervalReplyMenu renders the review-push interval picker.
func LearningPushIntervalReplyMenu(intervals []int) tgbotapi.ReplyKeyboardMarkup {
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
	rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(LearningButtonBack)))
	return buttonbuilder.RK(rows...)
}

// FormatPushIntervalButton renders one interval as reply-button text, e.g.
// "⏱ 60 min".
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
