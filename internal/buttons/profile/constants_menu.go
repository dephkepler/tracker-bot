package profile

// Inline callbacks.
const (
	ProfileCBEditLanguage = "profile:edit:language"
	ProfileCBEditTimeZone = "profile:edit:timezone"
	ProfileCBEditContact  = "profile:edit:contact"
	ProfileCBRefresh      = "profile:refresh"
)

// Language reply menu buttons. Deliberately NOT translated — each always
// shows its language's own native name regardless of the viewer's current
// interface language (like every other app's language picker), so these
// are plain constants here rather than i18n catalog entries. See
// handlers.languageCodeByButton for the button->stored-code mapping.
const (
	ProfileButtonLanguageRussian   = "🇷🇺 Русский"
	ProfileButtonLanguageEnglish   = "🇺🇸 English"
	ProfileButtonLanguageGerman    = "🇩🇪 Deutsch"
	ProfileButtonLanguageUkrainian = "🇺🇦 Українська"
	ProfileButtonLanguageArabian   = "🇸🇦 العربية"
)
