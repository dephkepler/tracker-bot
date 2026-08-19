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
// Phase 2: Track main screen, activity management, timer, archive. Reports
// (including the Track Today/Calendar/Period screens) is Phase 3 — still
// English until then, not an oversight.

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

// Track — nav buttons and labels (main screen, activity manage, archive,
// timer). Reply-keyboard buttons here (KeyTrackButtonBack/BackHome/
// ActivityActivate/ActivityDelete/TimerCreate/TimerDelete) are matched by
// the dispatcher via i18n.Key, same as the Phase 1 ones.
const (
	KeyTrackButtonSelectActivity = "track.button.select_activity"
	KeyTrackButtonCreateActivity = "track.button.create_activity"
	KeyTrackButtonViewReports    = "track.button.view_reports"
	KeyTrackButtonViewArchive    = "track.button.view_archive"

	// Back/Home reply buttons on Track/Timer/Archive screens render the
	// exact same text as everywhere else in the app — reuse
	// KeyCommonBack/KeyCommonHome rather than redefining them here (two
	// keys mapping to identical text would collide in the reverse index
	// Key() depends on; see TestCatalog_NoTextCollisionsWithinLanguage).
	KeyTrackButtonActivityActivate = "track.button.activity_activate"
	KeyTrackButtonActivityDelete   = "track.button.activity_delete"
	KeyTrackButtonTimerCreate      = "track.button.timer_create"
	KeyTrackButtonTimerDelete      = "track.button.timer_delete"

	KeyTrackLabelBack             = "track.label.back" // inline "↩️ Back"
	KeyTrackLabelOpenActivities   = "track.label.open_activities"
	KeyTrackLabelOpenArchive      = "track.label.open_archive"
	KeyTrackLabelCreateAnother    = "track.label.create_another"
	KeyTrackLabelArchiveSelected  = "track.label.archive_selected"
	KeyTrackLabelActiveActivities = "track.label.active_activities"
	KeyTrackLabelRestore          = "track.label.restore"
	KeyTrackLabelDeleteForever    = "track.label.delete_forever"
	KeyTrackLabelStopTimer        = "track.label.stop_timer"

	KeyTrackMainTitle           = "track.main.title"
	KeyTrackMainCurrentActivity = "track.main.current_activity"
	KeyTrackMainTodayTime       = "track.main.today_time"
	KeyTrackMainStreak          = "track.main.streak"
	KeyTrackMainTodayCount      = "track.main.today_count"
	KeyTrackMainProgress        = "track.main.progress" // "Progress: %s (%d%%, target %s)"
	KeyTrackLoadFailed          = "track.main.load_failed"

	KeyTrackCreatePrompt        = "track.create.prompt"
	KeyTrackCreateEmptyName     = "track.create.empty_name"
	KeyTrackCreateAlreadyExists = "track.create.already_exists"
	KeyTrackCreateFailed        = "track.create.failed"
	KeyTrackCreateConfirmed     = "track.create.confirmed" // "Created: %s"

	KeyTrackManageLoadFailed    = "track.manage.load_failed"
	KeyTrackManageEmpty         = "track.manage.empty"        // "No activities yet. Create one first."
	KeyTrackManageSelectTitle   = "track.manage.select_title" // "📂 Select Activity\n\nSelected: %d of %d"
	KeyTrackInvalidActivityID   = "track.invalid_activity_id" // shared by activity toggle and prompt-answer flows
	KeyTrackManageToggleFailed  = "track.manage.toggle_failed"
	KeyTrackManageRefreshFailed = "track.manage.refresh_failed"
	KeyTrackManageDeleteFailed  = "track.manage.delete_failed"
	KeyTrackManageDeleted       = "track.manage.deleted" // "🗑 Deleted: %d"

	KeyTrackArchiveFailed              = "track.archive.archive_failed"
	KeyTrackArchiveNoneSelected        = "track.archive.none_selected"
	KeyTrackArchived                   = "track.archive.archived" // "📦 Archived: %d"
	KeyTrackArchiveLoadFailed          = "track.archive.load_failed"
	KeyTrackArchiveEmpty               = "track.archive.empty"
	KeyTrackArchiveTitle               = "track.archive.title" // "🗄 Archive\n\nTotal archived: %d"
	KeyTrackArchiveInvalidItem         = "track.archive.invalid_item"
	KeyTrackArchiveRestoreFailed       = "track.archive.restore_failed"
	KeyTrackArchiveRestored            = "track.archive.restored" // "♻ Activity restored: %s"
	KeyTrackArchiveDeleteForeverFailed = "track.archive.delete_forever_failed"
	KeyTrackArchiveDeletedForever      = "track.archive.deleted_forever" // "🗑 Deleted forever: %s"

	// KeyTrackMinutesUnit is the abbreviated unit word used in timer button
	// text ("⏱ 15 min" / "⏱ 15 мин") — see FormatTimerButton/
	// ParseTimerButtonMinutes in internal/buttons/track.
	KeyTrackMinutesUnit             = "track.timer.minutes_unit"
	KeyTrackTimerPickerTitle        = "track.timer.picker_title"
	KeyTrackTimerCustomPrompt       = "track.timer.custom_prompt"
	KeyTrackTimerNoneToDelete       = "track.timer.none_to_delete"
	KeyTrackTimerPickToDelete       = "track.timer.pick_to_delete"
	KeyTrackTimerNotANumber         = "track.timer.not_a_number"
	KeyTrackTimerOutOfRange         = "track.timer.out_of_range" // "Interval must be between %d and %d minutes."
	KeyTrackTimerLimitReached       = "track.timer.limit_reached"
	KeyTrackTimerSaveFailed         = "track.timer.save_failed"
	KeyTrackTimerAdded              = "track.timer.added" // "✅ Custom timer added: %d min"
	KeyTrackTimerDeleteFailed       = "track.timer.delete_failed"
	KeyTrackTimerRemoved            = "track.timer.removed" // "🗑 Custom timer removed: %d min"
	KeyTrackTimerLoadSelectedFailed = "track.timer.load_selected_failed"
	KeyTrackTimerNoneSelected       = "track.timer.none_selected"
	KeyTrackTimerActivateFailed     = "track.timer.activate_failed"
	KeyTrackTimerActivated          = "track.timer.activated" // "✅ Timer activated: every %d min"
	KeyTrackTimerStopFailed         = "track.timer.stop_failed"
	KeyTrackTimerStopped            = "track.timer.stopped"

	KeyTrackPromptQuestion        = "track.prompt.question" // "What are you doing now?"
	KeyTrackPromptInvalidPayload  = "track.prompt.invalid_payload"
	KeyTrackPromptInvalidInterval = "track.prompt.invalid_interval"
	KeyTrackPromptSaveFailed      = "track.prompt.save_failed"
	KeyTrackPromptSaved           = "track.prompt.saved" // "Saved ✅\nActivity: %s\nTime: %s-%s (%d min)"
)
