package learning

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// the review button always opens the same picker; its label only hints whether reviews are running.
func LearningEntryInlineMenu(lang i18n.Lang, reviewsActive bool) tgbotapi.InlineKeyboardMarkup {
	reviewLabel := i18n.T(lang, i18n.KeyLearningButtonStartReviews)
	if reviewsActive {
		reviewLabel = i18n.T(lang, i18n.KeyLearningButtonManageReviews)
	}
	reviewBtn := buttonbuilder.IB(reviewLabel, LearningCBReviewOpen)
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonCreateCollection), LearningCBCreateCollection),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonWordBase), LearningCBWordBase),
		),
		buttonbuilder.IR(
			reviewBtn,
			buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonArchive), LearningCBArchiveOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonStatistics), LearningCBStats),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home"),
		),
	)
}

func LearningBackToMainInlineMenu(lang i18n.Lang) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), LearningCBBackMain)),
	)
}

func LearningWordBaseInlineMenu(lang i18n.Lang, items []models.LearningCollectionItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+1)
	for _, item := range items {
		label := i18n.T(lang, i18n.KeyLearningButtonToggleOffFmt, item.Name, item.WordCount)
		if item.Active {
			label = i18n.T(lang, i18n.KeyLearningButtonToggleOnFmt, item.Name, item.WordCount)
		}
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, fmt.Sprintf("%s%d", LearningCBCollectionOpen, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

// toggling a collection here takes effect immediately, with no separate "apply" step.
func LearningReviewPickInlineMenu(lang i18n.Lang, items []models.LearningCollectionItem, active bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)+2)
	for _, item := range items {
		label := i18n.T(lang, i18n.KeyLearningButtonToggleOffFmt, item.Name, item.WordCount)
		if item.Active {
			label = i18n.T(lang, i18n.KeyLearningButtonToggleOnFmt, item.Name, item.WordCount)
		}
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(label, fmt.Sprintf("%s%d", LearningCBReviewPickToggle, item.ID)),
		))
	}
	if active {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonStopReviews), LearningCBReviewStop)))
	} else {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonContinue), LearningCBReviewContinue)))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

func LearningCollectionDetailInlineMenu(lang i18n.Lang, collectionID int64, active bool, words []models.LearningWordItem) tgbotapi.InlineKeyboardMarkup {
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

	toggleLabel := i18n.T(lang, i18n.KeyLearningLabelExcludedReviews)
	if active {
		toggleLabel = i18n.T(lang, i18n.KeyLearningLabelIncludedInReviews)
	}
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(toggleLabel, fmt.Sprintf("%s%d", LearningCBCollectionToggle, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonAddWords), fmt.Sprintf("%s%d", LearningCBCollectionAddMore, collectionID)),
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonRename), fmt.Sprintf("%s%d", LearningCBCollectionRename, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(
		buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonArchiveThis), fmt.Sprintf("%s%d", LearningCBCollectionArchive, collectionID)),
	))
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), LearningCBWordBase)))
	return buttonbuilder.IK(rows...)
}

func LearningArchiveInlineMenu(lang i18n.Lang, items []models.LearningCollectionItem) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(items)*2+2)
	for _, item := range items {
		label := i18n.T(lang, i18n.KeyLearningArchiveItemFmt, item.Name, item.WordCount)
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelRestore), fmt.Sprintf("%s%d", LearningCBArchiveRestore, item.ID)),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyTrackLabelDeleteForever), fmt.Sprintf("%s%d", LearningCBArchiveDelete, item.ID)),
		))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonBack), LearningCBBackMain)))
	return buttonbuilder.IK(rows...)
}

func LearningReviewRevealInlineMenu(lang i18n.Lang, wordID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyLearningButtonShowAnswer), fmt.Sprintf("%s%d", LearningCBReviewReveal, wordID))),
	)
}

// grade drives the next SRS interval (models.LearningGrade), not just a correct/incorrect flag.
func LearningReviewGradeInlineMenu(lang i18n.Lang, wordID int64, again, hard, good, easy time.Duration) tgbotapi.InlineKeyboardMarkup {
	label := func(key string, d time.Duration) string {
		return fmt.Sprintf("%s (%s)", i18n.T(lang, key), FormatGradeDelay(d))
	}
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(label(i18n.KeyLearningButtonAgain, again), fmt.Sprintf("%s%d", LearningCBReviewAgain, wordID)),
			buttonbuilder.IB(label(i18n.KeyLearningButtonHard, hard), fmt.Sprintf("%s%d", LearningCBReviewHard, wordID)),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(label(i18n.KeyLearningButtonGood, good), fmt.Sprintf("%s%d", LearningCBReviewGood, wordID)),
			buttonbuilder.IB(label(i18n.KeyLearningButtonEasy, easy), fmt.Sprintf("%s%d", LearningCBReviewEasy, wordID)),
		),
	)
}

// unit abbreviations ("10m", "1d") are deliberately not translated, like track.FormatTimerButton.
func FormatGradeDelay(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()+0.5))
	}
	days := int(d.Hours()/24 + 0.5)
	if days < 1 {
		days = 1
	}
	return fmt.Sprintf("%dd", days)
}

func LearningWaitingReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancelX))),
	)
}

// unlike single-shot prompts (which auto-exit), this flow needs an explicit "Done" to leave.
func LearningAddWordsReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonDone))),
	)
}

const pushIntervalPrefix = "⏱ "

func LearningPushIntervalReplyMenu(lang i18n.Lang, intervals []int) tgbotapi.ReplyKeyboardMarkup {
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

// not translated, same reasoning as FormatGradeDelay/track.FormatTimerButton.
func FormatPushIntervalButton(minutes int) string {
	return fmt.Sprintf("%s%d min", pushIntervalPrefix, minutes)
}

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
