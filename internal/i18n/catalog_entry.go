package i18n

// catalogEntry holds the Home/Entry screen: the top-level nav buttons and
// the one-time welcome message.
var catalogEntry = map[string]entry{
	KeyEntryButtonProfile: {
		RU: "👤Мой профиль",
		EN: "👤My account",
		DE: "👤Mein Konto",
		UK: "👤Мій профіль",
		AR: "👤حسابي",
	},
	KeyEntryButtonTrack: {
		RU: "📈Трекер",
		EN: "📈Track",
		DE: "📈Tracker",
		UK: "📈Трекер",
		AR: "📈التتبع",
	},
	KeyEntryButtonLearning: {
		RU: "🧠Обучение",
		EN: "🧠Learning",
		DE: "🧠Lernen",
		UK: "🧠Навчання",
		AR: "🧠التعلّم",
	},
	KeyEntryButtonSubscription: {
		RU: "💳Подписка",
		EN: "💳Subscription",
		DE: "💳Abo",
		UK: "💳Підписка",
		AR: "💳الاشتراك",
	},
	KeyEntryWelcome: {
		RU: "Добро пожаловать в Tracker Bot!",
		EN: "Welcome to Tracker Bot!",
		DE: "Willkommen bei Tracker Bot!",
		UK: "Ласкаво просимо до Tracker Bot!",
		AR: "مرحبًا بك في Tracker Bot!",
	},
}
