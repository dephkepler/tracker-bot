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
//
// All button/label TEXT (main screen, activity management, timer, archive,
// reports) has moved to internal/i18n (see i18n.KeyTrack*). Only
// non-translatable technical constants remain here.

// TrackLabelArchiveItemPrefix is a plain emoji marker (no translatable
// text), used as-is in every language.
const TrackLabelArchiveItemPrefix = "📦 "

// TrackTimerActivatePrefix/TrackTimerDeletePrefix prefix a timer button's
// text (see FormatTimerButton/ParseTimerButtonMinutes) so the same
// "<prefix><N> <unit>" shape can mean either "activate this interval" (on
// the main timer screen) or "delete this custom interval" (on the
// delete-picker screen), depending on which screen the user is on. Plain
// emoji, not translated.
const (
	TrackTimerActivatePrefix = "⏱ "
	TrackTimerDeletePrefix   = "🗑 "
)

// BuiltInTimerIntervals are the always-available timer choices shown before
// any user-defined custom intervals.
var BuiltInTimerIntervals = []int{15, 30}
