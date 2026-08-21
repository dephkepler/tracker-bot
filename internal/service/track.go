package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"tracker-bot/internal/models"
	"tracker-bot/internal/repo"
	"tracker-bot/pkg/apptime"
)

type TrackerService interface {
	GetMainStats(ctx context.Context, userID int64, loc *time.Location) (models.MainStats, error)
	CreateActivity(ctx context.Context, userID int64, name, emoji string) (repo.Activity, error)
	ListActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error)
	ToggleSelectedActivity(ctx context.Context, userID, activityID int64) error
	DeleteSelectedActivities(ctx context.Context, userID int64) (int64, error)
	ListSelectedActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error)
	ListArchivedActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error)
	ArchiveSelectedActivities(ctx context.Context, userID int64) (int64, error)
	RestoreArchivedActivity(ctx context.Context, userID, activityID int64) error
	DeleteArchivedForever(ctx context.Context, userID, activityID int64) error
	// SetActivityTarget validates and stores one activity's daily time
	// target (models.MinActivityTargetMinutes..MaxActivityTargetMinutes).
	SetActivityTarget(ctx context.Context, userID, activityID int64, minutes int) error

	// ListReminderActivities returns activities currently receiving
	// reminders — distinct from ListSelectedActivities, which is the
	// Select-Activity screen's one-off batch-select scratchpad.
	ListReminderActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error)
	// AddSelectedToReminders adds whatever is currently checked on the
	// Select-Activity screen to the reminder set (existing members
	// untouched) and clears the checkboxes. Returns how many were staged.
	AddSelectedToReminders(ctx context.Context, userID int64) (int, error)
	RemoveFromReminders(ctx context.Context, userID, activityID int64) error
	GetTodayReport(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error)
	GetTodayReportBySelected(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error)
	GetPeriodReport(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) (models.ReportPeriodStats, error)
	GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location) (map[int]time.Duration, error)
	GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, loc *time.Location) ([]time.Time, []time.Duration, error)
	GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) ([]models.HourActivityDuration, error)
	GetTrackedDaysInRange(ctx context.Context, userID int64, from, to time.Time, loc *time.Location) ([]time.Time, error)
}

type trackerService struct {
	repo repo.TrackerRepository
}

func NewTrackerService(repo repo.TrackerRepository) TrackerService {
	return &trackerService{
		repo: repo,
	}
}

// resolveLoc defaults to apptime.Location when loc is nil.
func resolveLoc(loc *time.Location) *time.Location {
	if loc == nil {
		return apptime.Location
	}
	return loc
}

func (srv *trackerService) GetMainStats(ctx context.Context, userID int64, loc *time.Location) (models.MainStats, error) {
	if userID <= 0 {
		return models.MainStats{}, fmt.Errorf("main stats: invalid userID")
	}
	loc = resolveLoc(loc)
	tzName := loc.String()

	last, ok, err := srv.repo.GetLastTrackedActiveActivity(ctx, userID)
	if err != nil {
		return models.MainStats{}, err
	}
	if !ok {
		return models.MainStats{}, nil
	}

	total, err := srv.repo.GetTodayDurationByActivity(ctx, userID, last.ID, tzName)
	if err != nil {
		return models.MainStats{}, err
	}

	days, err := srv.repo.GetTrackedDaysDescByActivity(ctx, userID, last.ID, tzName)
	if err != nil {
		return models.MainStats{}, err
	}

	todayTrackedActivities, err := srv.repo.GetTodayTrackedActivitiesCount(ctx, userID, tzName)
	if err != nil {
		return models.MainStats{}, err
	}

	// Raw parts, not a joined label: this is also read by the dashboard API,
	// where an emoji glued to the name is not something a client can undo.
	// The Telegram screen joins them itself, in track.TrackingMenuText.
	return models.MainStats{
		CurrentActivityID:    last.ID,
		CurrentActivityName:  last.Name,
		CurrentActivityEmoji: last.Emoji,
		TodayTracked:         total,
		TodaySessions:        todayTrackedActivities,
		StreakDays:           calcStreakDays(days, apptime.NowIn(loc), loc),
		TargetMinutes:        last.TargetMinutes,
	}, nil
}

func (srv *trackerService) CreateActivity(ctx context.Context, userID int64, name, emoji string) (repo.Activity, error) {
	name = strings.TrimSpace(name)
	emoji = strings.TrimSpace(emoji)
	return srv.repo.Create(ctx, userID, name, emoji)
}

func (srv *trackerService) ListActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error) {
	activities, err := srv.repo.ListActive(ctx, userID)
	if err != nil {
		return nil, err
	}

	selectedIDs, err := srv.repo.SelectedListActive(ctx, userID)
	if err != nil {
		return nil, err
	}

	selected := make(map[int64]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = struct{}{}
	}

	items := make([]models.TrackActivityItem, 0, len(activities))
	for _, a := range activities {
		_, isSelected := selected[a.ID]
		items = append(items, models.TrackActivityItem{
			ID:            a.ID,
			Name:          a.Name,
			Emoji:         a.Emoji,
			Selected:      isSelected,
			TargetMinutes: a.TargetMinutes,
		})
	}

	return items, nil
}

func (srv *trackerService) ToggleSelectedActivity(ctx context.Context, userID, activityID int64) error {
	return srv.repo.ToggleSelectedActive(ctx, userID, activityID)
}

func (srv *trackerService) DeleteSelectedActivities(ctx context.Context, userID int64) (int64, error) {
	return srv.repo.DeleteSelected(ctx, userID)
}

func (srv *trackerService) ListSelectedActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error) {
	items, err := srv.ListActivities(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]models.TrackActivityItem, 0, len(items))
	for _, item := range items {
		if item.Selected {
			out = append(out, item)
		}
	}
	return out, nil
}

func (srv *trackerService) ListReminderActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error) {
	items, err := srv.ListActivities(ctx, userID)
	if err != nil {
		return nil, err
	}
	reminderIDs, err := srv.repo.ListReminderActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	inReminders := make(map[int64]struct{}, len(reminderIDs))
	for _, id := range reminderIDs {
		inReminders[id] = struct{}{}
	}

	out := make([]models.TrackActivityItem, 0, len(reminderIDs))
	for _, item := range items {
		if _, ok := inReminders[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out, nil
}

// AddSelectedToReminders adds whatever is currently checked on the
// Select-Activity screen to the reminder set (existing members untouched —
// this is additive) and then clears the checkboxes, so re-opening
// Select Activity starts blank instead of showing stale checkmarks that
// look like they still need clearing before adding something new.
func (srv *trackerService) AddSelectedToReminders(ctx context.Context, userID int64) (int, error) {
	selectedIDs, err := srv.repo.SelectedListActive(ctx, userID)
	if err != nil {
		return 0, err
	}
	if len(selectedIDs) == 0 {
		return 0, nil
	}
	if err := srv.repo.AddToReminders(ctx, userID, selectedIDs); err != nil {
		return 0, err
	}
	if err := srv.repo.ClearSelected(ctx, userID); err != nil {
		return 0, err
	}
	return len(selectedIDs), nil
}

func (srv *trackerService) RemoveFromReminders(ctx context.Context, userID, activityID int64) error {
	return srv.repo.RemoveFromReminders(ctx, userID, activityID)
}

func (srv *trackerService) ListArchivedActivities(ctx context.Context, userID int64) ([]models.TrackActivityItem, error) {
	activities, err := srv.repo.ListArchived(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]models.TrackActivityItem, 0, len(activities))
	for _, a := range activities {
		items = append(items, models.TrackActivityItem{
			ID:       a.ID,
			Name:     a.Name,
			Emoji:    a.Emoji,
			Selected: false,
		})
	}
	return items, nil
}

func (srv *trackerService) ArchiveSelectedActivities(ctx context.Context, userID int64) (int64, error) {
	return srv.repo.ArchiveSelected(ctx, userID)
}

func (srv *trackerService) RestoreArchivedActivity(ctx context.Context, userID, activityID int64) error {
	return srv.repo.RestoreArchived(ctx, userID, activityID)
}

func (srv *trackerService) DeleteArchivedForever(ctx context.Context, userID, activityID int64) error {
	return srv.repo.DeleteArchivedForever(ctx, userID, activityID)
}

func (srv *trackerService) SetActivityTarget(ctx context.Context, userID, activityID int64, minutes int) error {
	if minutes < models.MinActivityTargetMinutes || minutes > models.MaxActivityTargetMinutes {
		return models.ErrActivityTargetInvalid
	}
	return srv.repo.SetActivityTarget(ctx, userID, activityID, minutes)
}

// "today" means the user's local calendar day (loc), not UTC.
func (srv *trackerService) GetTodayReport(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error) {
	tzName := resolveLoc(loc).String()

	total, sessions, err := srv.repo.GetTodayStats(ctx, userID, tzName)
	if err != nil {
		return models.ReportTodayStats{}, err
	}

	acts, durs, cnts, err := srv.repo.GetTodayActivities(ctx, userID, tzName)
	if err != nil {
		return models.ReportTodayStats{}, err
	}

	top := make([]models.ActivityDurationStat, 0, len(acts))
	for i := range acts {
		top = append(top, models.ActivityDurationStat{
			ActivityID: acts[i].ID,
			Name:       acts[i].Name,
			Emoji:      acts[i].Emoji,
			Duration:   durs[i],
			Sessions:   cnts[i],
		})
	}

	return models.ReportTodayStats{
		TotalTracked:  total,
		TotalSessions: sessions,
		TopActivities: top,
	}, nil
}

func (srv *trackerService) GetTodayReportBySelected(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error) {
	report, err := srv.GetTodayReport(ctx, userID, loc)
	if err != nil {
		return models.ReportTodayStats{}, err
	}

	selectedIDs, err := srv.repo.SelectedListActive(ctx, userID)
	if err != nil {
		return models.ReportTodayStats{}, err
	}
	selected := make(map[int64]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = struct{}{}
	}

	filtered := make([]models.ActivityDurationStat, 0, len(report.TopActivities))
	var total time.Duration
	var sessions int
	for _, item := range report.TopActivities {
		if _, ok := selected[item.ActivityID]; !ok {
			continue
		}
		filtered = append(filtered, item)
		total += item.Duration
		sessions += item.Sessions
	}

	return models.ReportTodayStats{
		TotalTracked:  total,
		TotalSessions: sessions,
		TopActivities: filtered,
	}, nil
}

func (srv *trackerService) GetPeriodReport(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) (models.ReportPeriodStats, error) {
	tzName := resolveLoc(loc).String()

	acts, durs, cnts, total, sessions, err := srv.repo.GetPeriodActivities(ctx, userID, from, to, activityIDs)
	if err != nil {
		return models.ReportPeriodStats{}, err
	}

	items := make([]models.ActivityDurationStat, 0, len(acts))
	for i := range acts {
		items = append(items, models.ActivityDurationStat{
			ActivityID: acts[i].ID,
			Name:       acts[i].Name,
			Emoji:      acts[i].Emoji,
			Duration:   durs[i],
			Sessions:   cnts[i],
		})
	}
	months, monthDurs, err := srv.repo.GetPeriodMonthlyTotals(ctx, userID, from, to, activityIDs, tzName)
	if err != nil {
		return models.ReportPeriodStats{}, err
	}
	monthly := make([]models.MonthDurationStat, 0, len(months))
	for i := range months {
		monthly = append(monthly, models.MonthDurationStat{
			Month:    months[i],
			Duration: monthDurs[i],
		})
	}

	return models.ReportPeriodStats{
		From:          from,
		To:            to,
		TotalTracked:  total,
		TotalSessions: sessions,
		Activities:    items,
		Monthly:       monthly,
	}, nil
}

func (srv *trackerService) GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location) (map[int]time.Duration, error) {
	loc = resolveLoc(loc)
	return srv.repo.GetMonthDailyTotals(ctx, userID, month, activityIDs, loc, loc.String())
}

func (srv *trackerService) GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, loc *time.Location) ([]time.Time, []time.Duration, error) {
	return srv.repo.GetPeriodBuckets(ctx, userID, from, to, activityIDs, granularity, resolveLoc(loc).String())
}

func (srv *trackerService) GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) ([]models.HourActivityDuration, error) {
	return srv.repo.GetHourlyBucketsByActivity(ctx, userID, from, to, activityIDs, resolveLoc(loc).String())
}

func (srv *trackerService) GetTrackedDaysInRange(ctx context.Context, userID int64, from, to time.Time, loc *time.Location) ([]time.Time, error) {
	return srv.repo.GetTrackedDaysInRange(ctx, userID, from, to, resolveLoc(loc).String())
}

func calcStreakDays(days []time.Time, now time.Time, loc *time.Location) int {
	if len(days) == 0 {
		return 0
	}
	loc = resolveLoc(loc)
	daySet := make(map[string]struct{}, len(days))
	for _, d := range days {
		// Formatted, never converted. These values come from
		// `date_trunc('day', start_at AT TIME ZONE $n)::timestamp`, so each one
		// is already a local wall clock in loc, carrying time.UTC only because
		// that is what pgx attaches to a zoneless timestamp. Calling .In(loc)
		// here reinterpreted it as a UTC instant and shifted it — harmless east
		// of UTC, but west of it every key landed a day early, so today never
		// matched and the streak read 0. See TestCalcStreakDaysAcrossTimezones.
		daySet[d.Format("2006-01-02")] = struct{}{}
	}

	local := now.In(loc)
	cur := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	streak := 0
	for {
		key := cur.Format("2006-01-02")
		if _, ok := daySet[key]; !ok {
			break
		}
		streak++
		cur = cur.AddDate(0, 0, -1)
	}
	return streak
}
