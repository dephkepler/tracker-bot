package track

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

	// TrackCBHeatmapDay carries the tapped day as "<TrackCBHeatmapDay>2006-01-02".
	TrackCBHeatmapDay  = "track:heatmap:day:"
	TrackCBHeatmapBack = "track:heatmap:back"

	// TrackCBActivityTarget carries the activity id: "<TrackCBActivityTarget><id>".
	TrackCBActivityTarget = "track:activity:target:"
)

// most button/label text lives in internal/i18n (i18n.KeyTrack*); only non-translatable constants remain here

// plain emoji marker, not translated — same in every language
const TrackLabelArchiveItemPrefix = "📦 "

// same "<prefix><N> <unit>" shape means activate (timer screen) or delete (delete-picker), by context
const (
	TrackTimerActivatePrefix = "⏱ "
	TrackTimerDeletePrefix   = "🗑 "
)

var BuiltInTimerIntervals = []int{15, 30}
