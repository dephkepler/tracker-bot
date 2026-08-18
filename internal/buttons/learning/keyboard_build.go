package learning

import (
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

func LearningEntryInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonAddCollection, LearningCBAddCollection),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonRandomWords, LearningCBRandomWords),
			buttonbuilder.IB(LearningButtonSwitchCollection, LearningCBSwitchCollection),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(LearningButtonSummaryLearning, LearningCBSummaryLearning),
			buttonbuilder.IB(LearningButtonBaseWords, LearningCBBaseWords),
		),
		buttonbuilder.IR(
			buttonbuilder.IB("🏠 Home", "go_home"),
		),
	)
}

// Reply button menus

func LearningAddCollectionReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LearningButtonHelp), buttonbuilder.RB(LearningButtonHome)),
	)
}

func LearningAddWordsReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LearningButtonAddWord)),
		buttonbuilder.RR(buttonbuilder.RB(LearningButtonComplete), buttonbuilder.RB(LearningButtonBackHome)),
	)
}
