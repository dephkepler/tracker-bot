package onboarding

// button/label text here isn't localized yet — planned for i18n Phase 4 (see internal/i18n)
const (
	// CBGoto carries the step to show: "<CBGoto><0-4>".
	CBGoto = "onboarding:goto:"
	CBSkip = "onboarding:skip"
)

// StepCount is how many steps the tour has (0-indexed: 0..StepCount-1).
const StepCount = 5

const (
	ButtonNext   = "Next ▶"
	ButtonBack   = "◀ Back"
	ButtonSkip   = "Skip"
	ButtonFinish = "✅ Got it"
	ButtonHome   = "🏠 Home"

	ButtonGoTrack      = "📈 Go to Track"
	ButtonGoLearning   = "🧠 Go to Learning"
	ButtonGoChallenges = "🎯 Go to Challenges"
)
