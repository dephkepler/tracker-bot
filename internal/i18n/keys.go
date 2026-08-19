package i18n

// Key constants name every translatable string in catalog (see the
// catalog_*.go files). Other packages should always reference these
// constants rather than the raw dot-namespaced key strings — same reason
// button/callback constants are exported everywhere else in this codebase:
// one typo-proof source of truth.
//
// Filled in phase by phase (see doc/vps-deploy.md-style commit history —
// each phase covers one part of the UI and is deployed before the next).
// Phase 1: common strings + Entry (Home) screen + Profile screen.

// Common — shared across many screens, not tied to one feature.
const (
	KeyCommonBack           = "common.back"
	KeyCommonHome           = "common.home"
	KeyCommonCancel         = "common.cancel"
	KeyCommonCancelled      = "common.cancelled"
	KeyCommonAdmin          = "common.admin" // "👑 Admin" — identical button in Entry and Profile
	KeyCommonUnknownCommand = "common.unknown_command"
	KeyCommonHelpText       = "common.help_text"
	KeyCommonFallback       = "common.fallback" // catch-all "I don't know what to do with that"
	KeyCommonGenericError   = "common.generic_error"
	KeyCommonUseButtons     = "common.use_buttons"
)

// Entry (Home) screen.
const (
	KeyEntryButtonProfile      = "entry.button.profile"
	KeyEntryButtonTrack        = "entry.button.track"
	KeyEntryButtonLearning     = "entry.button.learning"
	KeyEntryButtonSubscription = "entry.button.subscription"
	KeyEntryWelcome            = "entry.welcome"
)

// Profile screen.
const (
	KeyProfileTitle          = "profile.title"
	KeyProfileLabelID        = "profile.label.id"
	KeyProfileLabelName      = "profile.label.name"
	KeyProfileLabelLanguage  = "profile.label.language" // shared with the "🌐 Language" button — same text
	KeyProfileLabelTimezone  = "profile.label.timezone"
	KeyProfileLabelEmail     = "profile.label.email"
	KeyProfileButtonTimezone = "profile.button.timezone"
	KeyProfileButtonContact  = "profile.button.contact"
	KeyProfileButtonRefresh  = "profile.button.refresh"
	KeyProfileLoadFailed     = "profile.load_failed"

	KeyProfileLanguagePrompt     = "profile.language.prompt"
	KeyProfileLanguageInvalid    = "profile.language.invalid"
	KeyProfileLanguageSaveFailed = "profile.language.save_failed"
	KeyProfileLanguageSaved      = "profile.language.saved"

	KeyProfileButtonShareLocation  = "profile.button.share_location"
	KeyProfileTimezonePrompt       = "profile.timezone.prompt"
	KeyProfileTimezoneInvalidTap   = "profile.timezone.invalid_tap"
	KeyProfileTimezoneLookupFailed = "profile.timezone.lookup_failed"
	KeyProfileTimezoneSaveFailed   = "profile.timezone.save_failed"
	KeyProfileTimezoneSaved        = "profile.timezone.saved"
)
