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

// TrackerService contains tracking use-cases used by handlers.
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
	GetTodayReport(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error)
	GetTodayReportBySelected(ctx context.Context, userID int64, loc *time.Location) (models.ReportTodayStats, error)
	GetPeriodReport(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) (models.ReportPeriodStats, error)
	GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location) (map[int]time.Duration, error)
	GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, loc *time.Location) ([]time.Time, []time.Duration, error)
	GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) ([]models.HourActivityDuration, error)
}

type trackerService struct {
	repo repo.TrackerRepository
}

// NewTrackerService creates tracking service.
func NewTrackerService(repo repo.TrackerRepository) TrackerService {
	return &trackerService{
		repo: repo,
	}
}

// resolveLoc defaults to apptime.Location when the caller has no specific
// user timezone in scope yet.
func resolveLoc(loc *time.Location) *time.Location {
	if loc == nil {
		return apptime.Location
	}
	return loc
}

// GetMainStats returns tracking home summary.
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

	streak := calcStreakDays(days, apptime.NowIn(loc), loc)
	currentName := last.Name
	if strings.TrimSpace(last.Emoji) != "" {
		currentName = last.Emoji + " " + last.Name
	}

	return models.MainStats{
		CurrentActivityName: currentName,
		TodayTracked:        total,
		TodaySessions:       todayTrackedActivities,
		StreakDays:          streak,
	}, nil
}

// CreateActivity validates and creates new activity.
func (srv *trackerService) CreateActivity(ctx context.Context, userID int64, name, emoji string) (repo.Activity, error) {
	name = strings.TrimSpace(name)
	emoji = strings.TrimSpace(emoji)
	return srv.repo.Create(ctx, userID, name, emoji)
}

// ListActivities returns active activities with selected flags.
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
			ID:       a.ID,
			Name:     a.Name,
			Emoji:    a.Emoji,
			Selected: isSelected,
		})
	}

	return items, nil
}

// ToggleSelectedActivity toggles activity selection state.
func (srv *trackerService) ToggleSelectedActivity(ctx context.Context, userID, activityID int64) error {
	return srv.repo.ToggleSelectedActive(ctx, userID, activityID)
}

// DeleteSelectedActivities removes currently selected activities.
func (srv *trackerService) DeleteSelectedActivities(ctx context.Context, userID int64) (int64, error) {
	return srv.repo.DeleteSelected(ctx, userID)
}

// ListSelectedActivities returns only selected active activities.
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

// ListArchivedActivities returns archived activities.
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

// ArchiveSelectedActivities moves selected activities to archive.
func (srv *trackerService) ArchiveSelectedActivities(ctx context.Context, userID int64) (int64, error) {
	return srv.repo.ArchiveSelected(ctx, userID)
}

// RestoreArchivedActivity moves one activity from archive back to active.
func (srv *trackerService) RestoreArchivedActivity(ctx context.Context, userID, activityID int64) error {
	return srv.repo.RestoreArchived(ctx, userID, activityID)
}

// DeleteArchivedForever permanently removes archived activity.
func (srv *trackerService) DeleteArchivedForever(ctx context.Context, userID, activityID int64) error {
	return srv.repo.DeleteArchivedForever(ctx, userID, activityID)
}

// GetTodayReport aggregates today's tracked durations and sessions, "today"
// meaning the user's own local calendar day.
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

// GetTodayReportBySelected returns today's report filtered by selected activities.
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

// GetPeriodReport aggregates report for date range and optional activity filter.
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

// GetMonthDailyTotals returns daily totals for given month.
func (srv *trackerService) GetMonthDailyTotals(ctx context.Context, userID int64, month time.Time, activityIDs []int64, loc *time.Location) (map[int]time.Duration, error) {
	loc = resolveLoc(loc)
	return srv.repo.GetMonthDailyTotals(ctx, userID, month, activityIDs, loc, loc.String())
}

// GetPeriodBuckets returns bucketed totals (hour/day/month).
func (srv *trackerService) GetPeriodBuckets(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, granularity string, loc *time.Location) ([]time.Time, []time.Duration, error) {
	return srv.repo.GetPeriodBuckets(ctx, userID, from, to, activityIDs, granularity, resolveLoc(loc).String())
}

// GetHourlyBucketsByActivity returns per-hour totals broken down by
// activity (see repo.TrackerRepository.GetHourlyBucketsByActivity).
func (srv *trackerService) GetHourlyBucketsByActivity(ctx context.Context, userID int64, from, to time.Time, activityIDs []int64, loc *time.Location) ([]models.HourActivityDuration, error) {
	return srv.repo.GetHourlyBucketsByActivity(ctx, userID, from, to, activityIDs, resolveLoc(loc).String())
}

func calcStreakDays(days []time.Time, now time.Time, loc *time.Location) int {
	if len(days) == 0 {
		return 0
	}
	loc = resolveLoc(loc)
	daySet := make(map[string]struct{}, len(days))
	for _, d := range days {
		daySet[d.In(loc).Format("2006-01-02")] = struct{}{}
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
