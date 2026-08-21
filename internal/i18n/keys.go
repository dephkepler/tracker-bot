package i18n

// Two keys with identical display text in the same language collide in the
// reverse i18n.Key() lookup the dispatcher uses to match reply-keyboard taps
// (see TestCatalog_NoTextCollisionsWithinLanguage) — reuse an existing key
// instead of adding a duplicate.

const (
	KeyCommonBack           = "common.back"
	KeyCommonHome           = "common.home"
	KeyCommonCancel         = "common.cancel"
	KeyCommonCancelled      = "common.cancelled"
	KeyCommonAdmin          = "common.admin"
	KeyCommonOnboarding     = "common.onboarding"
	KeyCommonUnknownCommand = "common.unknown_command"
	KeyCommonHelpText       = "common.help_text"
	KeyCommonFallback       = "common.fallback"
	KeyCommonGenericError   = "common.generic_error"
	KeyCommonUseButtons     = "common.use_buttons"

	KeyCommonCancelX = "common.cancel_x"
	KeyCommonDone    = "common.done"

	KeyCommonNameSingleLineInvalid = "common.name_single_line_invalid"
)

const (
	KeyEntryButtonProfile      = "entry.button.profile"
	KeyEntryButtonTrack        = "entry.button.track"
	KeyEntryButtonLearning     = "entry.button.learning"
	KeyEntryButtonRoadmap      = "entry.button.roadmap"
	KeyEntryButtonChallenge    = "entry.button.challenge"
	KeyEntryButtonSubscription = "entry.button.subscription"
	KeyEntryWelcome            = "entry.welcome"
)

// Challenges — 🎯 set a goal for N days and check off each day. Promoted
// from a Profile-screen inline sub-menu to a top-level main-menu feature
// (KeyEntryButtonChallenge above), and localized for the first time here —
// the module was previously entirely hardcoded English.
//
// Shared buttons are deliberately NOT redefined here: "◀ Back"/"🏠 Home"/
// "❌ Cancel"/"Cancelled." reuse KeyCommonBack/KeyCommonHome/
// KeyCommonCancelX/KeyCommonCancelled; the day-mark "✅ Done" button reuses
// KeyCommonDone; name validation reuses KeyCommonNameSingleLineInvalid;
// archive restore/delete reuse KeyTrackLabelRestore/
// KeyTrackLabelDeleteForever; "Failed to load archive." reuses
// KeyTrackArchiveLoadFailed; and the date-range calendar reuses Track's
// KeyTrackCalendarMon..Sun/Month01..12 (month/weekday names),
// KeyTrackLabelConfirmRange and KeyTrackLabelSelectEndDate — two keys
// rendering the same text would collide in the reverse index Key() depends
// on (see TestCatalog_NoTextCollisionsWithinLanguage).
const (
	KeyChallengeButtonCreate      = "challenge.button.create"
	KeyChallengeButtonArchive     = "challenge.button.archive"
	KeyChallengeButtonArchiveThis = "challenge.button.archive_this"
	KeyChallengeButtonSkip        = "challenge.button.skip"

	KeyChallengeListTitleFmt   = "challenge.list.title_fmt" // "🎯 *Challenges* — %d active"
	KeyChallengeListEmpty      = "challenge.list.empty"
	KeyChallengeListLoadFailed = "challenge.list.load_failed"
	KeyChallengeItemLabelFmt   = "challenge.list.item_label_fmt"    // "🎯 %s — %d%% (%d/%d)"
	KeyChallengeArchiveItemFmt = "challenge.archive.item_label_fmt" // "📦 %s — %d/%d done"

	KeyChallengeNotFound     = "challenge.not_found"
	KeyChallengeLoadFailed   = "challenge.load_failed"
	KeyChallengeDayNotFound  = "challenge.day.not_found"
	KeyChallengeGridTitleFmt = "challenge.grid.title_fmt"

	// The day-detail screen: status line, then the proportions "donut",
	// trend strip, and streak — all appended above the existing Done/Skip
	// buttons, not replacing them.
	KeyChallengeDayStatusUnmarked    = "challenge.day.status_unmarked"
	KeyChallengeDayStatusDoneText    = "challenge.day.status_done"
	KeyChallengeDayStatusSkippedText = "challenge.day.status_skipped"
	KeyChallengeDayHeaderFmt         = "challenge.day.header_fmt"      // "🎯 *%s*\n\n%s — %s."
	KeyChallengeDayProportionsFmt    = "challenge.day.proportions_fmt" // "✅ Done %d%%   ❌ Skipped %d%%   🔲 Left %d%%"
	KeyChallengeDayTrendLabelFmt     = "challenge.day.trend_label_fmt" // "📈 Trend (last %d days):"
	KeyChallengeDayStreakFmt         = "challenge.day.streak_fmt"      // "🔥 Current streak: %d days   🏆 Best: %d days"
	KeyChallengeDayMarkPrompt        = "challenge.day.mark_prompt"

	KeyChallengeArchiveTitleFmt = "challenge.archive.title_fmt"
	KeyChallengeArchiveEmpty    = "challenge.archive.empty"

	KeyChallengeCreateNamePrompt     = "challenge.create.name_prompt"
	KeyChallengeCreateRangeIntro     = "challenge.create.range_intro"
	KeyChallengeCreateCalendarHeader = "challenge.create.calendar_header"
	KeyChallengeCreateExists         = "challenge.create.exists"
	KeyChallengeCreateInvalidRange   = "challenge.create.invalid_range"
	KeyChallengeCreateFailed         = "challenge.create.failed"
	KeyChallengeCreatedFmt           = "challenge.create.created_fmt"

	KeyChallengePushTextFmt       = "challenge.push.text_fmt"
	KeyChallengePushMarkedDone    = "challenge.push.marked_done"
	KeyChallengePushMarkedSkipped = "challenge.push.marked_skipped"
)

const (
	KeyProfileTitle          = "profile.title"
	KeyProfileLabelID        = "profile.label.id"
	KeyProfileLabelName      = "profile.label.name"
	KeyProfileLabelLanguage  = "profile.label.language"
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

const (
	KeyTrackButtonSelectActivity = "track.button.select_activity"
	KeyTrackButtonCreateActivity = "track.button.create_activity"
	KeyTrackButtonViewReports    = "track.button.view_reports"
	KeyTrackButtonViewArchive    = "track.button.view_archive"

	KeyTrackButtonActivityActivate = "track.button.activity_activate"
	KeyTrackButtonActivityDelete   = "track.button.activity_delete"
	KeyTrackButtonTimerCreate      = "track.button.timer_create"
	KeyTrackButtonTimerDelete      = "track.button.timer_delete"

	KeyTrackLabelBack             = "track.label.back"
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
	KeyTrackMainProgress        = "track.main.progress"
	KeyTrackLoadFailed          = "track.main.load_failed"

	// Per-activity daily target — the 🎯 button on the activity list
	// (internal/buttons/track/keyboard_build.go's TrackActivitiesInlineMenu)
	// and its numeric-input flow.
	KeyTrackActivityTargetPromptFmt    = "track.activity.target_prompt_fmt"
	KeyTrackActivityTargetInvalid      = "track.activity.target_invalid"
	KeyTrackActivityTargetSavedFmt     = "track.activity.target_saved_fmt"
	KeyTrackActivityTargetSaveFailed   = "track.activity.target_save_failed"
	KeyTrackActivityTargetButtonSetFmt = "track.activity.target_button_set_fmt"
	KeyTrackActivityTargetButtonUnset  = "track.activity.target_button_unset"

	KeyTrackCreatePrompt        = "track.create.prompt"
	KeyTrackCreatePromptBlocked = "track.create.prompt_blocked"
	KeyTrackCreateEmptyName     = "track.create.empty_name"
	KeyTrackCreateAlreadyExists = "track.create.already_exists"
	KeyTrackCreateFailed        = "track.create.failed"
	KeyTrackCreateConfirmed     = "track.create.confirmed"

	KeyTrackManageMenuClosed    = "track.manage.menu_closed"
	KeyTrackManageLoadFailed    = "track.manage.load_failed"
	KeyTrackManageEmpty         = "track.manage.empty"
	KeyTrackManageSelectTitle   = "track.manage.select_title"
	KeyTrackInvalidActivityID   = "track.invalid_activity_id"
	KeyTrackManageToggleFailed  = "track.manage.toggle_failed"
	KeyTrackManageRefreshFailed = "track.manage.refresh_failed"
	KeyTrackManageDeleteFailed  = "track.manage.delete_failed"
	KeyTrackManageDeleted       = "track.manage.deleted"

	KeyTrackArchiveFailed              = "track.archive.archive_failed"
	KeyTrackArchiveNoneSelected        = "track.archive.none_selected"
	KeyTrackArchived                   = "track.archive.archived"
	KeyTrackArchiveLoadFailed          = "track.archive.load_failed"
	KeyTrackArchiveEmpty               = "track.archive.empty"
	KeyTrackArchiveTitle               = "track.archive.title"
	KeyTrackArchiveInvalidItem         = "track.archive.invalid_item"
	KeyTrackArchiveRestoreFailed       = "track.archive.restore_failed"
	KeyTrackArchiveRestored            = "track.archive.restored"
	KeyTrackArchiveDeleteForeverFailed = "track.archive.delete_forever_failed"
	KeyTrackArchiveDeletedForever      = "track.archive.deleted_forever"

	// KeyTrackMinutesUnit is machine-parsed by ParseTimerButtonMinutes
	// (internal/buttons/track) — the abbreviation format matters, not just
	// how it displays.
	KeyTrackMinutesUnit             = "track.timer.minutes_unit"
	KeyTrackTimerPickerTitle        = "track.timer.picker_title"
	KeyTrackTimerCustomPrompt       = "track.timer.custom_prompt"
	KeyTrackTimerPromptBlocked      = "track.timer.prompt_blocked"
	KeyTrackTimerNoneToDelete       = "track.timer.none_to_delete"
	KeyTrackTimerPickToDelete       = "track.timer.pick_to_delete"
	KeyTrackTimerNotANumber         = "track.timer.not_a_number"
	KeyTrackTimerOutOfRange         = "track.timer.out_of_range"
	KeyTrackTimerLimitReached       = "track.timer.limit_reached"
	KeyTrackTimerSaveFailed         = "track.timer.save_failed"
	KeyTrackTimerAdded              = "track.timer.added"
	KeyTrackTimerDeleteFailed       = "track.timer.delete_failed"
	KeyTrackTimerRemoved            = "track.timer.removed"
	KeyTrackTimerLoadSelectedFailed = "track.timer.load_selected_failed"
	KeyTrackTimerNoneSelected       = "track.timer.none_selected"
	KeyTrackTimerActivateFailed     = "track.timer.activate_failed"
	KeyTrackTimerActivated          = "track.timer.activated"
	KeyTrackTimerStopFailed         = "track.timer.stop_failed"
	KeyTrackTimerStopped            = "track.timer.stopped"

	KeyTrackPromptQuestion        = "track.prompt.question"
	KeyTrackPromptInvalidPayload  = "track.prompt.invalid_payload"
	KeyTrackPromptInvalidInterval = "track.prompt.invalid_interval"
	KeyTrackPromptSaveFailed      = "track.prompt.save_failed"
	KeyTrackPromptSaved           = "track.prompt.saved"
)

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
	KeyTrackLabelRangePrefix        = "track.label.range_prefix"
	KeyTrackLabelConfirmRange       = "track.label.confirm_range"
	KeyTrackLabelSelectEndDate      = "track.label.select_end_date"
	// KeyTrackCalendarCancel is deliberately distinct from KeyCommonCancel:
	// the calendar's Cancel has always been plain "Cancel", no ✖️ — reusing
	// the common key would silently change that text.
	KeyTrackCalendarCancel = "track.calendar.cancel"
	// KeyTrackCalendarMonth is a static filler label between ◀/▶, not an
	// actual month name — those are KeyTrackCalendarMonth01..12 below.
	KeyTrackCalendarMonth = "track.calendar.month"

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
	KeyTrackHeatmapDaysTracked = "track.reports.heatmap.days_tracked"
	KeyTrackHeatmapHint        = "track.reports.heatmap.hint"

	KeyTrackHeatmapDayTrackedHeader = "track.reports.heatmap.day.tracked_header"
	KeyTrackHeatmapDayNoActivity    = "track.reports.heatmap.day.no_activity"
	KeyTrackHeatmapDayActivityLine  = "track.reports.heatmap.day.activity_line"
	KeyTrackHeatmapDayReviewsHeader = "track.reports.heatmap.day.reviews_header"
	KeyTrackHeatmapDayNoReviews     = "track.reports.heatmap.day.no_reviews"
	KeyTrackHeatmapDayReviewLine    = "track.reports.heatmap.day.review_line"

	KeyTrackTodayChartLoadFailed = "track.reports.today_chart.load_failed"
	KeyTrackTodayChartEmpty      = "track.reports.today_chart.empty"
	KeyTrackTodayChartTitle      = "track.reports.today_chart.title"
	// KeyTrackTodayChartActivityLine has no session-count arg, unlike
	// KeyTrackPeriodChartActivityLine below — don't reuse one's formatting
	// code for the other, the arg counts differ.
	KeyTrackTodayChartActivityLine = "track.reports.today_chart.activity_line"

	KeyTrackTodaySelectTitle = "track.reports.today_select.title"

	KeyTrackPeriodMenuLoadFailed = "track.reports.period_menu.load_failed"
	KeyTrackPeriodMenuTitle      = "track.reports.period_menu.title"

	KeyTrackPeriodTextFailed        = "track.reports.period_text.failed"
	KeyTrackPeriodTextTitle         = "track.reports.period_text.title"
	KeyTrackPeriodRangeLine         = "track.reports.period.range_line"
	KeyTrackPeriodScopeSelected     = "track.reports.period.scope_selected"
	KeyTrackPeriodScopeAll          = "track.reports.period.scope_all"
	KeyTrackPeriodTotalsLine        = "track.reports.period.totals_line"
	KeyTrackPeriodNoSessions        = "track.reports.period.no_sessions"
	KeyTrackPeriodTextActivityLine  = "track.reports.period_text.activity_line"
	KeyTrackPeriodChartFailed       = "track.reports.period_chart.failed"
	KeyTrackPeriodChartEmpty        = "track.reports.period_chart.empty"
	KeyTrackPeriodChartTitle        = "track.reports.period_chart.title"
	KeyTrackPeriodChartActivityLine = "track.reports.period_chart.activity_line"

	KeyTrackCalendarPickTitle = "track.reports.calendar.pick_title"

	KeyTrackGranularityByMonths   = "track.reports.granularity.by_months"
	KeyTrackGranularityByDays     = "track.reports.granularity.by_days"
	KeyTrackGranularityByHours    = "track.reports.granularity.by_hours"
	KeyTrackGranularityBucketLine = "track.reports.granularity.bucket_line"
	KeyTrackPeriodRangeInvalidFmt = "track.reports.period.invalid_format"
	KeyTrackPeriodRangeSetConfirm = "track.reports.period.range_set"
	KeyTrackCalendarPickBothDays  = "track.reports.calendar.pick_both_days"
)

const (
	KeyLearningButtonCreateCollection = "learning.button.create_collection"
	KeyLearningButtonStartReviews     = "learning.button.start_reviews"
	KeyLearningButtonManageReviews    = "learning.button.manage_reviews"
	KeyLearningButtonStopReviews      = "learning.button.stop_reviews"
	KeyLearningButtonArchive          = "learning.button.archive"
	KeyLearningButtonStatistics       = "learning.button.statistics"
	KeyLearningButtonWordBase         = "learning.button.word_base"
	KeyLearningButtonContinue         = "learning.button.continue"
	KeyLearningButtonAddWords         = "learning.button.add_words"
	KeyLearningButtonRename           = "learning.button.rename"
	KeyLearningButtonArchiveThis      = "learning.button.archive_this"
	KeyLearningButtonToggleOnFmt      = "learning.button.toggle_on_fmt"
	KeyLearningButtonToggleOffFmt     = "learning.button.toggle_off_fmt"
	KeyLearningButtonShowAnswer       = "learning.button.show_answer"
	KeyLearningButtonAgain            = "learning.button.grade_again"
	KeyLearningButtonHard             = "learning.button.grade_hard"
	KeyLearningButtonGood             = "learning.button.grade_good"
	KeyLearningButtonEasy             = "learning.button.grade_easy"
	KeyLearningLabelIncludedInReviews = "learning.label.included_in_reviews"
	KeyLearningLabelExcludedReviews   = "learning.label.excluded_from_reviews"
	KeyLearningArchiveItemFmt         = "learning.label.archive_item_fmt"

	KeyLearningMenuTitle           = "learning.menu.title"
	KeyLearningMenuTotalWords      = "learning.menu.total_words"
	KeyLearningMenuDueToday        = "learning.menu.due_today"
	KeyLearningMenuLearned         = "learning.menu.learned"
	KeyLearningMenuStreak          = "learning.menu.streak"
	KeyLearningMenuReviewsActive   = "learning.menu.reviews_active"
	KeyLearningMenuReviewsInactive = "learning.menu.reviews_inactive"
	KeyLearningLoadFailed          = "learning.load_failed"

	KeyLearningReviewPickManageTitle = "learning.review_pick.manage_title"
	KeyLearningReviewPickEmptyTitle  = "learning.review_pick.empty_title"
	KeyLearningReviewPickTitle       = "learning.review_pick.title"
	KeyLearningReviewPickNeedOne     = "learning.review_pick.need_one"

	KeyLearningStatsTitle          = "learning.stats.title"
	KeyLearningStatsAccuracy       = "learning.stats.accuracy"
	KeyLearningStatsNoReviews      = "learning.stats.no_reviews"
	KeyLearningStatsByCollection   = "learning.stats.by_collection"
	KeyLearningStatsCollectionLine = "learning.stats.collection_line"
	KeyLearningStatsLoadFailed     = "learning.stats.load_failed"

	KeyLearningRenamePrompt = "learning.rename.prompt"
	KeyLearningRenamed      = "learning.rename.confirmed"
	KeyLearningRenameFailed = "learning.rename.failed"

	KeyLearningWordBaseTitle      = "learning.word_base.title"
	KeyLearningWordBaseEmpty      = "learning.word_base.empty"
	KeyLearningCollectionDetail   = "learning.collection.detail_title"
	KeyLearningCollectionNotFound = "learning.collection.not_found"
	KeyLearningCollectionsFailed  = "learning.collection.load_failed"
	KeyLearningWordsLoadFailed    = "learning.word.load_failed"

	KeyLearningCreatePrompt    = "learning.create.prompt"
	KeyLearningCreateNotAList  = "learning.create.not_a_list"
	KeyLearningCreateTooShort  = "learning.create.too_short"
	KeyLearningCreateTooLong   = "learning.create.too_long"
	KeyLearningCreateExists    = "learning.create.exists"
	KeyLearningCreateFailed    = "learning.create.failed"
	KeyLearningCreateConfirmed = "learning.create.confirmed"

	KeyLearningAddWordsPromptFirst = "learning.add_words.prompt_first"
	KeyLearningAddWordsPromptMore  = "learning.add_words.prompt_more"
	KeyLearningAddWordsNoneParsed  = "learning.add_words.none_parsed"
	KeyLearningAddWordsFailed      = "learning.add_words.failed"
	KeyLearningAddWordsAdded       = "learning.add_words.added"
	KeyLearningAddWordsSkipped     = "learning.add_words.skipped"
	KeyLearningAddWordsDoneNotice  = "learning.add_words.done_notice"

	KeyLearningArchiveTitle = "learning.archive.title"
	KeyLearningArchiveEmpty = "learning.archive.empty"

	KeyLearningReviewIntervalPrompt = "learning.review.interval_prompt"
	KeyLearningReviewActivateFailed = "learning.review.activate_failed"
	KeyLearningReviewActivated      = "learning.review.activated"
	KeyLearningReviewCardTitle      = "learning.review.card_title"
	KeyLearningReviewRevealed       = "learning.review.revealed"
	KeyLearningReviewLearned        = "learning.review.graded_learned"
	// KeyLearningReviewCorrect is shown for the "Good" grade — the key name
	// doesn't match the Again/Hard/Good/Easy grading labels above.
	KeyLearningReviewCorrect     = "learning.review.graded_correct"
	KeyLearningReviewHardConfirm = "learning.review.graded_hard"
	KeyLearningReviewEasyConfirm = "learning.review.graded_easy"
	// KeyLearningReviewMissed is shown for the "Again" grade — same
	// name/label mismatch as KeyLearningReviewCorrect above.
	KeyLearningReviewMissed = "learning.review.graded_missed"
	// KeyLearningReviewMissedMinutes/KeyLearningReviewHardConfirmMinutes
	// replace Missed/HardConfirm when the next review lands under a day
	// away (see service.reviewDelay) — otherwise the message would
	// misleadingly say "tomorrow" for a review due in 10-15 minutes.
	KeyLearningReviewMissedMinutes      = "learning.review.graded_missed_minutes"
	KeyLearningReviewHardConfirmMinutes = "learning.review.graded_hard_minutes"
)

// Roadmap — a three-level learning plan: a goal ("reach mid-level"), the
// technologies feeding it, and a checklist of cards under each technology.
//
// Shared buttons are deliberately not redefined here: Back/Home/Done/Cancel
// reuse the KeyCommon* ones, name validation reuses
// KeyCommonNameSingleLineInvalid, archive restore/delete reuse
// KeyTrackLabelRestore/KeyTrackLabelDeleteForever, and "Rename" reuses
// KeyLearningButtonRename — two keys rendering the same text collide in the
// reverse index Key() depends on (TestCatalog_NoTextCollisionsWithinLanguage).
const (
	KeyRoadmapButtonCreateGoal      = "roadmap.button.create_goal"
	KeyRoadmapButtonAddTech         = "roadmap.button.add_tech"
	KeyRoadmapButtonGoals           = "roadmap.button.goals"
	KeyRoadmapButtonOrphans         = "roadmap.button.orphans"
	KeyRoadmapButtonAssignGoal      = "roadmap.button.assign_goal"
	KeyRoadmapButtonStartReminders  = "roadmap.button.start_reminders"
	KeyRoadmapButtonManageReminders = "roadmap.button.manage_reminders"
	KeyRoadmapButtonStopReminders   = "roadmap.button.stop_reminders"
	KeyRoadmapButtonArchive         = "roadmap.button.archive"
	KeyRoadmapButtonProgress        = "roadmap.button.progress"
	KeyRoadmapButtonAddCards        = "roadmap.button.add_cards"
	KeyRoadmapButtonSetCriteria     = "roadmap.button.set_criteria"
	KeyRoadmapButtonArchiveGoal     = "roadmap.button.archive_goal"
	KeyRoadmapButtonArchiveThis     = "roadmap.button.archive_this"
	KeyRoadmapButtonSkipCriteria    = "roadmap.button.skip_criteria"

	KeyRoadmapGoalItemFmt             = "roadmap.label.goal_item_fmt"    // "🎯 %s — %d/%d (%d%%)"
	KeyRoadmapButtonToggleOnFmt       = "roadmap.button.toggle_on_fmt"   // "🟢 %s — %d/%d"
	KeyRoadmapButtonToggleOffFmt      = "roadmap.button.toggle_off_fmt"  // "⚪ %s — %d/%d"
	KeyRoadmapArchiveItemFmt          = "roadmap.label.archive_item_fmt" // "📦 %s — %d/%d"
	KeyRoadmapArchiveGoalItemFmt      = "roadmap.label.archive_goal_fmt"
	KeyRoadmapLabelIncludedInReminder = "roadmap.label.included_in_reminders"
	KeyRoadmapLabelExcludedReminder   = "roadmap.label.excluded_from_reminders"

	KeyRoadmapMenuTitle             = "roadmap.menu.title"
	KeyRoadmapMenuTotalGoals        = "roadmap.menu.total_goals"
	KeyRoadmapMenuTotalRoadmaps     = "roadmap.menu.total_roadmaps"
	KeyRoadmapMenuTotalCards        = "roadmap.menu.total_cards"
	KeyRoadmapMenuDone              = "roadmap.menu.done"
	KeyRoadmapMenuPending           = "roadmap.menu.pending"
	KeyRoadmapMenuRemindersActive   = "roadmap.menu.reminders_active"
	KeyRoadmapMenuRemindersInactive = "roadmap.menu.reminders_inactive"
	KeyRoadmapLoadFailed            = "roadmap.load_failed"

	KeyRoadmapGoalsTitle      = "roadmap.goals.title"
	KeyRoadmapGoalsEmpty      = "roadmap.goals.empty"
	KeyRoadmapGoalsLoadFailed = "roadmap.goals.load_failed"

	KeyRoadmapGoalDetailTitle  = "roadmap.goal.detail_title"
	KeyRoadmapGoalDetailCounts = "roadmap.goal.detail_counts"
	KeyRoadmapGoalDetailNoTech = "roadmap.goal.detail_no_tech"
	KeyRoadmapGoalDetailHint   = "roadmap.goal.detail_hint"
	KeyRoadmapGoalCreatePrompt = "roadmap.goal.create_prompt"
	KeyRoadmapGoalExists       = "roadmap.goal.exists"
	KeyRoadmapGoalCreateFailed = "roadmap.goal.create_failed"
	KeyRoadmapGoalCreated      = "roadmap.goal.created"
	KeyRoadmapGoalLimitReached = "roadmap.goal.limit_reached"
	KeyRoadmapGoalNotFound     = "roadmap.goal.not_found"
	KeyRoadmapGoalRenamePrompt = "roadmap.goal.rename_prompt"
	KeyRoadmapGoalRenamed      = "roadmap.goal.renamed"
	KeyRoadmapGoalRenameFailed = "roadmap.goal.rename_failed"

	KeyRoadmapListTitle        = "roadmap.list.title"
	KeyRoadmapListEmpty        = "roadmap.list.empty"
	KeyRoadmapOrphansTitle     = "roadmap.orphans.title"
	KeyRoadmapOrphansEmpty     = "roadmap.orphans.empty"
	KeyRoadmapDetailTitle      = "roadmap.detail.title"
	KeyRoadmapDetailCriteria   = "roadmap.detail.criteria"
	KeyRoadmapDetailNoCriteria = "roadmap.detail.no_criteria"
	KeyRoadmapDetailHint       = "roadmap.detail.hint"
	KeyRoadmapDetailNoCards    = "roadmap.detail.no_cards"
	KeyRoadmapNotFound         = "roadmap.not_found"
	KeyRoadmapCardsLoadFailed  = "roadmap.cards.load_failed"

	KeyRoadmapCreatePrompt    = "roadmap.create.prompt"
	KeyRoadmapCreateExists    = "roadmap.create.exists"
	KeyRoadmapCreateFailed    = "roadmap.create.failed"
	KeyRoadmapCreateConfirmed = "roadmap.create.confirmed"
	KeyRoadmapLimitReached    = "roadmap.create.limit_reached"

	KeyRoadmapAssignPrompt  = "roadmap.assign.prompt"
	KeyRoadmapAssignNoGoals = "roadmap.assign.no_goals"
	KeyRoadmapAssigned      = "roadmap.assign.done"
	KeyRoadmapAssignFailed  = "roadmap.assign.failed"

	KeyRoadmapCriteriaPrompt  = "roadmap.criteria.prompt"
	KeyRoadmapCriteriaTooLong = "roadmap.criteria.too_long"
	KeyRoadmapCriteriaSaved   = "roadmap.criteria.saved"
	KeyRoadmapCriteriaSkipped = "roadmap.criteria.skipped"
	KeyRoadmapCriteriaFailed  = "roadmap.criteria.failed"

	KeyRoadmapRenamePrompt = "roadmap.rename.prompt"
	KeyRoadmapRenamed      = "roadmap.rename.confirmed"
	KeyRoadmapRenameFailed = "roadmap.rename.failed"

	KeyRoadmapAddCardsPromptFirst = "roadmap.add_cards.prompt_first"
	KeyRoadmapAddCardsPromptMore  = "roadmap.add_cards.prompt_more"
	KeyRoadmapAddCardsNoneParsed  = "roadmap.add_cards.none_parsed"
	KeyRoadmapAddCardsFailed      = "roadmap.add_cards.failed"
	KeyRoadmapAddCardsAdded       = "roadmap.add_cards.added"
	KeyRoadmapAddCardsSkipped     = "roadmap.add_cards.skipped"
	KeyRoadmapAddCardsDoneNotice  = "roadmap.add_cards.done_notice"

	KeyRoadmapArchiveTitle    = "roadmap.archive.title"
	KeyRoadmapArchiveEmpty    = "roadmap.archive.empty"
	KeyRoadmapArchiveGoalsHdr = "roadmap.archive.goals_header"
	KeyRoadmapArchiveTechHdr  = "roadmap.archive.tech_header"

	KeyRoadmapPushIntervalPrompt = "roadmap.push.interval_prompt"
	KeyRoadmapPushActivateFailed = "roadmap.push.activate_failed"
	KeyRoadmapPushActivated      = "roadmap.push.activated"
	KeyRoadmapPushNeedOne        = "roadmap.push.need_one"

	KeyRoadmapDigestTitle       = "roadmap.digest.title"
	KeyRoadmapDigestRoadmapLine = "roadmap.digest.roadmap_line"
	KeyRoadmapDigestEmpty       = "roadmap.digest.empty"

	KeyRoadmapStatsTitle        = "roadmap.stats.title"
	KeyRoadmapStatsGoalLine     = "roadmap.stats.goal_line"
	KeyRoadmapStatsRoadmapLine  = "roadmap.stats.roadmap_line"
	KeyRoadmapStatsCriteriaLine = "roadmap.stats.criteria_line"
	KeyRoadmapStatsNoGoalHeader = "roadmap.stats.no_goal_header"
	KeyRoadmapStatsEmpty        = "roadmap.stats.empty"
	KeyRoadmapStatsLoadFailed   = "roadmap.stats.load_failed"

	// Roadmap AI. Every one of these only ever renders when a provider is
	// configured — with AI off the buttons are not drawn at all, so there is
	// no "feature unavailable" string among them beyond the race guard.
	KeyRoadmapButtonAIPlan    = "roadmap.ai.button.plan"
	KeyRoadmapButtonAIPaste   = "roadmap.ai.button.paste"
	KeyRoadmapButtonAIQuiz    = "roadmap.ai.button.quiz"
	KeyRoadmapButtonQuizDone  = "roadmap.ai.button.quiz_done"
	KeyRoadmapAIWorking       = "roadmap.ai.working"
	KeyRoadmapAIQuizWorking   = "roadmap.ai.quiz_working"
	KeyRoadmapAIFailed        = "roadmap.ai.failed"
	KeyRoadmapAIEmpty         = "roadmap.ai.empty"
	KeyRoadmapAIDisabled      = "roadmap.ai.disabled"
	KeyRoadmapAIPlanAddedFmt  = "roadmap.ai.plan_added_fmt"
	KeyRoadmapAIRejectedFmt   = "roadmap.ai.rejected_fmt"
	KeyRoadmapAIPastePrompt   = "roadmap.ai.paste_prompt"
	KeyRoadmapAIQuizPromptFmt = "roadmap.ai.quiz_prompt_fmt"
	KeyRoadmapAIQuizCorrect   = "roadmap.ai.quiz_correct"
	KeyRoadmapAIQuizPartial   = "roadmap.ai.quiz_partial"
	KeyRoadmapAIQuizWrong     = "roadmap.ai.quiz_wrong"
	KeyRoadmapAIQuizGradeFmt  = "roadmap.ai.quiz_grade_fmt"
	KeyRoadmapAIDigestHintFmt = "roadmap.ai.digest_hint_fmt"
)
