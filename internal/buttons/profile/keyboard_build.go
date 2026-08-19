package profile

import (
	"tracker-bot/internal/buttons/admin"
	"tracker-bot/internal/i18n"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

// ProfileEntryInlineMenu renders the profile screen's inline menu in lang.
// isAdmin adds a row that opens the admin screen — pass
// handlers.Module.IsAdmin(ctx), never hardcode true. Labels here are inline
// (callback_data-routed, not text-matched), so translating them carries no
// routing risk — unlike reply-keyboard buttons.
func ProfileEntryInlineMenu(lang i18n.Lang, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileLabelLanguage), ProfileCBEditLanguage),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonTimezone), ProfileCBEditTimeZone),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonContact), ProfileCBEditContact),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonRefresh), ProfileCBRefresh),
		),
	}
	if isAdmin {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonAdmin), admin.CBOpen)))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonHome), "go_home")))
	return buttonbuilder.IK(rows...)
}

// Reply button menus

// ProfileLanguageManageReplyMenu renders the language picker. The 5
// language buttons themselves are not translated (see constants_menu.go);
// only the Cancel button is.
func ProfileLanguageManageReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageEnglish)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageRussian), buttonbuilder.RB(ProfileButtonLanguageGerman)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageUkrainian), buttonbuilder.RB(ProfileButtonLanguageArabian)),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancel))),
	)
}

// ProfileLocationReplyMenu prompts the user to share their location so the
// bot can detect their real timezone (see pkg/geotz).
func ProfileLocationReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RBLocation(i18n.T(lang, i18n.KeyProfileButtonShareLocation))),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancel))),
	)
}
