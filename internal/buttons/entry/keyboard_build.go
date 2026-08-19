package entry

import (
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Reply button menus

// EntryReplyMenu renders the home screen keyboard. isAdmin adds a row with
// EntryButtonAdmin — pass handlers.Module.IsAdmin(ctx), never hardcode true.
func EntryReplyMenu(isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	rows := [][]tgbotapi.KeyboardButton{
		buttonbuilder.RR(
			buttonbuilder.RB(EntryButtonProfile),
			buttonbuilder.RB(EntryButtonTrack),
		),
		buttonbuilder.RR(
			buttonbuilder.RB(EntryButtonLearning),
			buttonbuilder.RB(EntryButtonSubscription),
		),
	}
	if isAdmin {
		rows = append(rows, buttonbuilder.RR(buttonbuilder.RB(EntryButtonAdmin)))
	}
	return buttonbuilder.RK(rows...)
}
