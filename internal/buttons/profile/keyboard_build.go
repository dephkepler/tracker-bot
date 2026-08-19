package profile

import (
	"tracker-bot/internal/buttons/challenge"
	"tracker-bot/internal/i18n"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

// ProfileEntryInlineMenu renders the profile screen's inline menu in lang.
// Labels here are inline (callback_data-routed, not text-matched), so
// translating them carries no routing risk — unlike reply-keyboard buttons.
//
// No Home/Admin row here: both are already one tap away via the bottom
// reply keyboard (Home always visible, Admin its own reply button for
// admins), so this space is used for "🎯 Challenges" instead — see
// internal/buttons/challenge.
func ProfileEntryInlineMenu(lang i18n.Lang, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	_ = isAdmin // kept for signature stability; Admin no longer shown here.
	rows := [][]tgbotapi.InlineKeyboardButton{
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileLabelLanguage), ProfileCBEditLanguage),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonTimezone), ProfileCBEditTimeZone),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonContact), ProfileCBEditContact),
			buttonbuilder.IB(i18n.T(lang, i18n.KeyProfileButtonRefresh), ProfileCBRefresh),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonChallenges), challenge.CBBackList),
		),
	}
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
