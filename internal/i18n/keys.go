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
// Phase 2: Track main screen, activity management, timer, archive.
// Phase 3: Reports (Today/Calendar/Period).

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
	KeyTrackCreatePromptBlocked = "track.create.prompt_blocked" // typed while a nav button was expected instead
	KeyTrackCreateEmptyName     = "track.create.empty_name"
	KeyTrackCreateAlreadyExists = "track.create.already_exists"
	KeyTrackCreateFailed        = "track.create.failed"
	KeyTrackCreateConfirmed     = "track.create.confirmed" // "Created: %s"

	KeyTrackManageMenuClosed    = "track.manage.menu_closed"
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
	KeyTrackTimerPromptBlocked      = "track.timer.prompt_blocked" // typed while a nav button was expected instead
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

// Reports (Today/Calendar/Period).
const (
	KeyTrackButtonToday    = "track.button.today"
	KeyTrackButtonCalendar = "track.button.calendar"
	KeyTrackButtonHeatmap  = "track.button.heatmap"

	KeyTrackLabelBackToReports      = "track.label.back_to_reports"
	KeyTrackLabelSelectedActivities = "track.label.selected_activities"
	KeyTrackLabelTextReport         = "track.label.text_report"
	KeyTrackLabelChartReport        = "track.label.chart_report"
	KeyTrackLabelSelectActivities   = "track.label.select_activities"
	KeyTrackLabelBuildChart         = "track.label.build_chart"
	KeyTrackLabelRangePrefix        = "track.label.range_prefix" // "🗓 Range: %s"
	KeyTrackLabelConfirmRange       = "track.label.confirm_range"
	KeyTrackLabelSelectEndDate      = "track.label.select_end_date"
	// KeyTrackCalendarCancel is distinct from KeyCommonCancel — the
	// calendar's cancel button has always been plain "Cancel", no ✖️, so
	// reusing the common key would visibly change it.
	KeyTrackCalendarCancel = "track.calendar.cancel"
	KeyTrackCalendarMonth  = "track.calendar.month" // static placeholder between ◀/▶

	KeyTrackCalendarMon = "track.calendar.weekday.mon"
	KeyTrackCalendarTue = "track.calendar.weekday.tue"
	KeyTrackCalendarWed = "track.calendar.weekday.wed"
	KeyTrackCalendarThu = "track.calendar.weekday.thu"
	KeyTrackCalendarFri = "track.calendar.weekday.fri"
	KeyTrackCalendarSat = "track.calendar.weekday.sat"
	KeyTrackCalendarSun = "track.calendar.weekday.sun"

	KeyTrackCalendarMonth01 = "track.calendar.month_name.01"
	KeyTrackCalendarMonth02 = "track.calendar.month_name.02"
	KeyTrackCalendarMonth03 = "track.calendar.month_name.03"
	KeyTrackCalendarMonth04 = "track.calendar.month_name.04"
	KeyTrackCalendarMonth05 = "track.calendar.month_name.05"
	KeyTrackCalendarMonth06 = "track.calendar.month_name.06"
	KeyTrackCalendarMonth07 = "track.calendar.month_name.07"
	KeyTrackCalendarMonth08 = "track.calendar.month_name.08"
	KeyTrackCalendarMonth09 = "track.calendar.month_name.09"
	KeyTrackCalendarMonth10 = "track.calendar.month_name.10"
	KeyTrackCalendarMonth11 = "track.calendar.month_name.11"
	KeyTrackCalendarMonth12 = "track.calendar.month_name.12"

	KeyTrackReportsHubTitle = "track.reports.hub_title"

	KeyTrackHeatmapTitle       = "track.reports.heatmap.title"
	KeyTrackHeatmapLegend      = "track.reports.heatmap.legend"
	KeyTrackHeatmapDaysTracked = "track.reports.heatmap.days_tracked" // "%d of %d days tracked"

	KeyTrackTodayChartLoadFailed = "track.reports.today_chart.load_failed"
	KeyTrackTodayChartEmpty      = "track.reports.today_chart.empty"
	KeyTrackTodayChartTitle      = "track.reports.today_chart.title"
	// KeyTrackTodayChartActivityLine has no session count, unlike
	// KeyTrackPeriodChartActivityLine below — today's chart never shows it.
	KeyTrackTodayChartActivityLine = "track.reports.today_chart.activity_line" // "%s\n%s %s (%s)\n\n"

	KeyTrackTodaySelectTitle = "track.reports.today_select.title"

	KeyTrackPeriodMenuLoadFailed = "track.reports.period_menu.load_failed"
	KeyTrackPeriodMenuTitle      = "track.reports.period_menu.title" // "📅 Period Report\nSelected: %d activities\nRange: %s"

	KeyTrackPeriodTextFailed        = "track.reports.period_text.failed"
	KeyTrackPeriodTextTitle         = "track.reports.period_text.title"
	KeyTrackPeriodRangeLine         = "track.reports.period.range_line" // "Range: %s..%s\n"
	KeyTrackPeriodScopeSelected     = "track.reports.period.scope_selected"
	KeyTrackPeriodScopeAll          = "track.reports.period.scope_all"
	KeyTrackPeriodTotalsLine        = "track.reports.period.totals_line" // "Total: %s\nSessions: %d\n\n"
	KeyTrackPeriodNoSessions        = "track.reports.period.no_sessions"
	KeyTrackPeriodTextActivityLine  = "track.reports.period_text.activity_line" // "%d) %s - %s (%s, %d)\n"
	KeyTrackPeriodChartFailed       = "track.reports.period_chart.failed"
	KeyTrackPeriodChartEmpty        = "track.reports.period_chart.empty"
	KeyTrackPeriodChartTitle        = "track.reports.period_chart.title"
	KeyTrackPeriodChartActivityLine = "track.reports.period_chart.activity_line" // "%s\n%s %s (%s, %d)\n\n"

	KeyTrackCalendarPickTitle = "track.reports.calendar.pick_title" // "📅 Pick period days\nFrom: %s\nTo: %s"

	KeyTrackGranularityByMonths   = "track.reports.granularity.by_months"
	KeyTrackGranularityByDays     = "track.reports.granularity.by_days"
	KeyTrackGranularityByHours    = "track.reports.granularity.by_hours"
	KeyTrackGranularityBucketLine = "track.reports.granularity.bucket_line" // "- %s: %s\n"
	KeyTrackPeriodRangeInvalidFmt = "track.reports.period.invalid_format"
	KeyTrackPeriodRangeSetConfirm = "track.reports.period.range_set" // "Range set: %s..%s"
	KeyTrackCalendarPickBothDays  = "track.reports.calendar.pick_both_days"
)
