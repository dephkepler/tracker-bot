package profile

// Inline callbacks.
const (
	ProfileCBEditLanguage = "profile:edit:language"
	ProfileCBEditTimeZone = "profile:edit:timezone"
	ProfileCBEditContact  = "profile:edit:contact"
	ProfileCBRefresh      = "profile:refresh"
)

// Inline menu buttons.
const (
	ProfileButtonEditLanguage = "🌐 Language"
	ProfileButtonEditTimeZone = "📍 Time zone"
	ProfileButtonEditContact  = "📧 Contact"
	ProfileButtonRefresh      = "🔁 Refresh"
	// ProfileButtonAdmin only ever gets rendered for the configured admin
	// user (see ProfileEntryInlineMenu's isAdmin param) — never shown to
	// regular users, and access is re-checked server-side on the callback
	// regardless (handlers.Module.IsAdmin), not just gated by visibility.
	ProfileButtonAdmin = "👑 Admin"
)

// Language reply menu buttons.
const (
	ProfileButtonLanguageRussian   = "🇷🇺 Русский"
	ProfileButtonLanguageEnglish   = "🇺🇸 English"
	ProfileButtonLanguageGerman    = "🇩🇪 Deutsch"
	ProfileButtonLanguageUkrainian = "🇺🇦 Українська"
	ProfileButtonLanguageArabian   = "🇸🇦 العربية"
)

// Time zone reply menu buttons.
const (
	ProfileButtonShareLocation = "📍 Share location"
	ProfileButtonCancel        = "✖️ Cancel"
)

// Profile screen labels.
const (
	ProfileUIMainTitle    = "👤 Profile"
	ProfileUIMainID       = "🛜 ID:"
	ProfileUIMainName     = "👤 Name:"
	ProfileUIMainLanguage = "🌐 Language"
	ProfileUIMainTimeZone = "📍 Time zone:"
	ProfileUIMainEmail    = "📧 Email:"
)
