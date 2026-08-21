package entry

import (
	"tracker-bot/internal/i18n"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// isAdmin adds the "👑 Admin" row — pass handlers.Module.IsAdmin(ctx), never hardcode true.
func EntryReplyMenu(lang i18n.Lang, isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	rows := [][]tgbotapi.KeyboardButton{
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonProfile)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonTrack)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonLearning)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonRoadmap)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonChallenge)),
		),
	}
	if isAdmin {
		rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonAdmin))))
	}
	return buttonbuilder.RK(rows...)
}
