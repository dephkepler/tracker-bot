package profile

import (
	"fmt"
	"tracker-bot/internal/buttons/onboarding"
	"tracker-bot/internal/buttons/subscription"
	"tracker-bot/internal/i18n"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline labels are callback-routed, not text-matched, so translating them carries no routing risk (unlike reply buttons).
// miniAppURL empty hides the dashboard row entirely. That is the state before
// the Mini App is registered with BotFather, and a button that opens nothing is
// worse than no button.
func ProfileEntryInlineMenu(lang i18n.Lang, isAdmin bool, miniAppURL string) tgbotapi.InlineKeyboardMarkup {
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
			// Subscription's main-menu reply button was retired in favor of
			// Challenges — reusing the same label key here, just rendered
			// inline instead of on the reply keyboard.
			buttonbuilder.IB(i18n.T(lang, i18n.KeyEntryButtonSubscription), subscription.SubscriptionCBOpen),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(i18n.T(lang, i18n.KeyCommonOnboarding), fmt.Sprintf("%s0", onboarding.CBGoto)),
		),
	}
	if miniAppURL != "" {
		// A URL button, not a callback: Telegram opens the link itself, so this
		// never reaches the dispatcher and needs no routing.
		rows = append(rows, buttonbuilder.IR(
			buttonbuilder.IBURL(i18n.T(lang, i18n.KeyProfileButtonDashboard), miniAppURL),
		))
	}
	return buttonbuilder.IK(rows...)
}

// The 5 language buttons are not translated (see constants_menu.go); only Cancel is.
func ProfileLanguageManageReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageEnglish)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageRussian), buttonbuilder.RB(ProfileButtonLanguageGerman)),
		buttonbuilder.RR(buttonbuilder.RB(ProfileButtonLanguageUkrainian), buttonbuilder.RB(ProfileButtonLanguageArabian)),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancel))),
	)
}

func ProfileLocationReplyMenu(lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RBLocation(i18n.T(lang, i18n.KeyProfileButtonShareLocation))),
		buttonbuilder.RR(buttonbuilder.RB(i18n.T(lang, i18n.KeyCommonCancel))),
	)
}
