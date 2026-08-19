package profile

import (
	"tracker-bot/internal/buttons/admin"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

// ProfileEntryInlineMenu renders the profile screen's inline menu. isAdmin
// adds a row that opens the admin screen — pass handlers.Module.IsAdmin(ctx),
// never hardcode true.
func ProfileEntryInlineMenu(isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		buttonbuilder.IR(
			buttonbuilder.IB(ProfileButtonEditLanguage, ProfileCBEditLanguage),
			buttonbuilder.IB(ProfileButtonEditTimeZone, ProfileCBEditTimeZone),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(ProfileButtonEditContact, ProfileCBEditContact),
			buttonbuilder.IB(ProfileButtonRefresh, ProfileCBRefresh),
		),
	}
	if isAdmin {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(ProfileButtonAdmin, admin.CBOpen)))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB("🏠 Home", "go_home")))
	return buttonbuilder.IK(rows...)
}

// Reply button menus

func ProfileLanguageManageReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageEnglish)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageRussian), buttonbuilder.RB(ProfileButtonLanguageGerman)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageUkrainian), buttonbuilder.RB(ProfileButtonLanguageArabian)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonCancel)),
	)
}

// ProfileLocationReplyMenu prompts the user to share their location so the
// bot can detect their real timezone (see pkg/geotz).
func ProfileLocationReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RBLocation(ProfileButtonShareLocation)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonCancel)),
	)
}
