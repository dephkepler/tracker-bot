package subscription

import (
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Inline button menus

func SubscriptionEntryInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(SubscriptionButtonTariffPlans, SubscriptionCBTariffPlans),
			buttonbuilder.IB(SubscriptionButtonFreePlan, SubscriptionCBFreePlan),
		),
		buttonbuilder.IR(
			buttonbuilder.IB(SubscriptionButtonSupport, SubscriptionCBSupport),
			buttonbuilder.IB(SubscriptionButtonPaymentChange, SubscriptionCBPaymentChange),
		),
		buttonbuilder.IR(
			buttonbuilder.IB("🏠 Home", "go_home"),
		),
	)
}

// Reply button menus
