package entry

import (
	"tracker-bot/internal/i18n"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Reply button menus

// EntryReplyMenu renders the home screen keyboard in lang. isAdmin adds a
// row with the "👑 Admin" button — pass handlers.Module.IsAdmin(ctx), never
// hardcode true.
func EntryReplyMenu(lang i18n.Lang, isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	rows := [][]tgbotapi.KeyboardButton{
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonProfile)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonTrack)),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonLearning)),
			buttonbuilder.RB(i18n.T(lang, i18n.KeyEntryButtonSubscription)),
		),
	}
	if isAdmin {
		rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonAdmin))))
	}
	return buttonbuilder.RK(rows...)
}
