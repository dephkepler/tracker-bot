package subscription

const (
	SubscriptionCBTariffPlans   = "subscription:tariff:plans"
	SubscriptionCBFreePlan      = "subscription:free:plan"
	SubscriptionCBSupport       = "subscription:support"
	SubscriptionCBPaymentChange = "subscription:payment:change"
	// SubscriptionCBOpen is Profile's inline entry point into the
	// subscription screen — Subscription no longer has its own main-menu
	// reply button (see internal/buttons/entry/keyboard_build.go).
	SubscriptionCBOpen = "subscription:open"
)

const (
	SubscriptionButtonTariffPlans   = "🗓 Tariff plans"
	SubscriptionButtonFreePlan      = "🎁 Free"
	SubscriptionButtonSupport       = "🛫 Support"
	SubscriptionButtonPaymentChange = "💳 Change payment"
)

const (
	SubscriptionUIMainTitle      = "💳 Subscription"
	SubscriptionUIMainTariffPlan = "🗓 Tariff plan:"
	SubscriptionUIMainDaysEnd    = "🕐 Days end:"
	SubscriptionUIMainMessage    = "To subscribe, go to: 🗓 Tariff plans"
)
