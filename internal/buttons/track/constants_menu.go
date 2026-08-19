package track

// ---------------------------------------------------------------------
// Inline callbacks (actions)
const (
	TrackCBActivitySelect         = "track:activity:select"
	TrackCBActivityCreate         = "track:activity:create"
	TrackCBPromptActivity         = "track:prompt:activity:"
	TrackCBPromptStopTimer        = "track:prompt:stop"
	TrackCBReportSummary          = "track:report:summary"
	TrackCBArchiveOpen            = "track:archive:open"
	TrackCBArchiveSelected        = "track:archive:selected"
	TrackCBArchiveRestore         = "track:archive:restore:"
	TrackCBArchiveDelete          = "track:archive:delete:"
	TrackCBArchiveToActive        = "track:archive:to_active"
	TrackCBOpenActivities         = "track:open:activities"
	TrackCBCreateAnother          = "track:create:another"
	TrackCBOpenArchive            = "track:open:archive"
	TrackCBActivityReportOpen     = "track:activity:report"
	TrackCBReportsHub             = "track:report:hub"
	TrackCBReportsToday           = "track:report:today"
	TrackCBReportsBackHub         = "track:report:back:hub"
	TrackCBReportsTodayBySelected = "track:report:today:selected"
	TrackCBReportsTodaySelToggle  = "track:report:today:selected:toggle:"
	TrackCBReportsTodaySelBuild   = "track:report:today:selected:build"
	TrackCBReportsPeriodOpen      = "track:report:period:open"
	TrackCBReportsPeriodToggle    = "track:report:period:toggle:"
	TrackCBReportsPeriodSetRange  = "track:report:period:set_range"
	TrackCBReportsPeriodText      = "track:report:period:text"
	TrackCBReportsPeriodChart     = "track:report:period:chart"
	TrackCBReportsCalPrev         = "track:report:cal:prev"
	TrackCBReportsCalNext         = "track:report:cal:next"
	TrackCBReportsCalPrevYear     = "track:report:cal:prev_year"
	TrackCBReportsCalNextYear     = "track:report:cal:next_year"
	TrackCBReportsCalPick         = "track:report:cal:pick:"
	TrackCBReportsCalDone         = "track:report:cal:done"
	TrackCBReportsCalCancel       = "track:report:cal:cancel"
	TrackCBReportsCalThisMonth    = "track:report:cal:this_month"
	TrackCBReportsCalThisYear     = "track:report:cal:this_year"
)

// ---------------------------------------------------------------------
// Buttons (Inline + Reply)

// Entry inline menu buttons
const (
	TrackButtonSelectActivity = "📂 Activities"
	TrackButtonCreateActivity = "➕ New Activity"
	TrackButtonExitTracking   = "⏹ Stop Tracking"
	TrackButtonViewReports    = "📈 Reports"
	TrackButtonViewArchive    = "🗄 Archive"
)

// Shared inline labels
const (
	TrackLabelBack               = "↩️ Back"
	TrackLabelBackToReports      = "↩️ Back to Reports"
	TrackLabelOpenActivities     = "📂 Open Activities"
	TrackLabelOpenArchive        = "🗄 Open Archive"
	TrackLabelCreateAnother      = "➕ Create Another"
	TrackLabelArchiveSelected    = "🛒 Archive selected"
	TrackLabelActiveActivities   = "📂 Active activities"
	TrackLabelRestore            = "♻ Restore"
	TrackLabelDeleteForever      = "🗑 Delete forever"
	TrackLabelSelectedActivities = "Selected activities"
	TrackLabelTextReport         = "📄 Text report"
	TrackLabelChartReport        = "📉 Chart report"
	TrackLabelSelectActivities   = "🧩 Select activities"
	TrackLabelBuildChart         = "✅ Build chart"
	TrackLabelStopTimer          = "⏹ Stop Timer"
	TrackLabelRangePrefix        = "🗓 Range: "
	TrackLabelConfirmRange       = "✅ Confirm range"
	TrackLabelSelectEndDate      = "Select end date"
	TrackLabelCancel             = "Cancel"
	TrackLabelMonth              = "Month"
	TrackLabelMon                = "Mo"
	TrackLabelTue                = "Tu"
	TrackLabelWed                = "We"
	TrackLabelThu                = "Th"
	TrackLabelFri                = "Fr"
	TrackLabelSat                = "Sa"
	TrackLabelSun                = "Su"
	TrackLabelArchiveItemPrefix  = "📦 "
)

// Common reply buttons
const (
	TrackButtonToday    = "📊 Today"
	TrackButtonPeriod   = "📅 Calendar"
	TrackButtonBack     = "◀ Back"
	TrackButtonBackHome = "🏠 Home"
)

// Report reply menu buttons
const (
	TrackButtonReportPeriod = "📅 Period"
	TrackButtonReportWeek   = "🗓 Week"
	TrackButtonReportExport = "📤 Export"
	TrackButtonReportDelete = "🗑 Delete"
)

// Activity manage reply menu buttons
const (
	TrackButtonActivityActivate = "📳 Activate"
	TrackButtonActivityDelete   = "🗑 Delete"
)

// Timer picker (reply menu)
const (
	TrackButtonTimerCreate = "➕ Custom Timer"
	TrackButtonTimerDelete = "🗑 Delete Timer"

	// TrackTimerActivatePrefix/TrackTimerDeletePrefix prefix a timer button's
	// text (see FormatTimerButton/ParseTimerButtonMinutes) so the same
	// "<prefix><N> min" shape can mean either "activate this interval" (on
	// the main timer screen) or "delete this custom interval" (on the
	// delete-picker screen), depending on which screen the user is on.
	TrackTimerActivatePrefix = "⏱ "
	TrackTimerDeletePrefix   = "🗑 "
)

// BuiltInTimerIntervals are the always-available timer choices shown before
// any user-defined custom intervals.
var BuiltInTimerIntervals = []int{15, 30}

// ---------------------------------------------------------------------
// Track UI texts (titles/labels shown inside messages)

// Main screen
const (
	TrackUIMainTitle                = "📈 Tracking"
	TrackUIMainLabelCurrentActivity = "📌 Current activity:"
	TrackUIMainLabelTodayTime       = "⏱ Tracked today:"
	TrackUIMainLabelStreak          = "🔥 Streak:"
	TrackUIMainLabelTodayCount      = "✅ Sessions today:"
)

// Activity report screen
const (
	TrackUIReportTitle                = "📌 Activity report"
	TrackUIReportLabelStartDate       = "📅 Started:"
	TrackUIReportLabelConsecutiveDays = "📈 Streak:"
	TrackUIReportLabelTodayTimeTotal  = "⏱ Today total:"
	TrackUIReportLabelAvgDailyTime    = "📊 Daily average:"
	TrackUIReportLabelTodayDate       = "🗓 Date:"
)

// ---------------------------------------------------------------------
// Messages (plain texts, not labels/titles)
const (
	TrackMsgActivityListTitle     = "📂 Select Activity"
	TrackMsgActivityListConfirmed = "📂 Activated Activities:"
	TrackMsgTimerPickerTitle      = "⏱ Select tracking interval:"
	TrackMsgCustomTimerPrompt     = "Enter custom interval in minutes (1-360):"
)
