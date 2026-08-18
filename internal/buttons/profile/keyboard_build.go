package profile

import (
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

func ProfileEntryInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(ProfileButtonEditLanguage, ProfileCBEditLanguage),
			buttonbuilder.IB(ProfileButtonEditTimeZone, ProfileCBEditTimeZone),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(ProfileButtonEditContact, ProfileCBEditContact),
			buttonbuilder.IB(ProfileButtonRefresh, ProfileCBRefresh),
		),
		buttonbuilder.IR(
			buttonbuilder.IB("🏠 Home", "go_home"),
		),
	)
}

// Reply button menus

func ProfileLanguageManageReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageEnglish)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageRussian), buttonbuilder.RB(ProfileButtonLanguageGerman)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageUkrainian), buttonbuilder.RB(ProfileButtonLanguageArabian)),
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
