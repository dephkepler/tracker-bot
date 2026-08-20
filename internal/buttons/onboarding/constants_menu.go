package onboarding

// Inline callbacks. Text is not yet localized (see internal/i18n) — this
// module is planned for i18n Phase 4, like the other recently-added ones.
const (
	// CBGoto carries the step to show: "<CBGoto><0-4>".
	CBGoto = "onboarding:goto:"
	CBSkip = "onboarding:skip"
)

// StepCount is how many steps the tour has (0-indexed: 0..StepCount-1).
const StepCount = 5

// Inline menu buttons.
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
