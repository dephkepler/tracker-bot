package entry

import "tracker-bot/internal/i18n"

func WelcomeText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyEntryWelcome)
}

// deliberately reuses the "Home" button's own key (i18n.KeyCommonHome) — same word, same concept.
func HomeMenuText(lang i18n.Lang) string {
	return i18n.T(lang, i18n.KeyCommonHome)
}
