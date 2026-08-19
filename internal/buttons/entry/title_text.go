package entry

import "tracker-bot/internal/i18n"

// WelcomeText greets a user on their very first /start.
func WelcomeText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyEntryWelcome)
}

// HomeMenuText is shown every other time the user lands on the home screen.
// Deliberately the same text/key as the "🏠 Home" button (see
// i18n.KeyCommonHome) — they're the same word for the same concept.
func HomeMenuText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyCommonHome)
}
