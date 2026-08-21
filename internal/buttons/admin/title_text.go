package admin

import (
	"fmt"
	"tracker-bot/internal/models"
)

func OverviewStatsText(s models.AdminOverviewStats) string {
	return fmt.Sprintf(
		"📊 Overview\n\n👥 Users: %d\n⏱ Active track timers: %d\n🎲 Active review pushes: %d\n🗂 Activities: %d\n📚 Learning collections: %d\n🧠 Learning words: %d",
		s.TotalUsers, s.ActiveTrackTimers, s.ActiveReviewPushes, s.TotalActivities, s.TotalCollections, s.TotalLearningWords,
	)
}

func UserDetailText(d models.AdminUserDetail) string {
	return fmt.Sprintf(
		"👤 User #%d\n\nTelegram: %s (id `%d`)\nLanguage: %s\nTimezone: %s\nRegistered: %s\n\n🗂 Activities: %d\n📚 Learning collections: %d (%d words)\n⏱ Track timer: %s\n🎲 Reviews: %s",
		d.DBID, usernameLabel(d.UserName), d.TgUserID,
		valueOrDash(d.Language), valueOrDash(d.TimeZone), d.CreatedAt.Format("2006-01-02"),
		d.ActivitiesCount, d.CollectionsCount, d.LearningWords,
		activeLabel(d.TrackTimerActive), activeLabel(d.ReviewsActive),
	)
}

func UserDeleteConfirmText(d models.AdminUserDetail) string {
	return fmt.Sprintf("⚠️ Really delete user #%d (%s)?\n\nThis permanently removes their profile, activities, sessions, and learning data. This cannot be undone.", d.DBID, usernameLabel(d.UserName))
}

func UserDeletedText(dbID int64) string {
	return fmt.Sprintf("🗑 User #%d deleted.", dbID)
}

const BroadcastPromptText = "📢 Send the message you want to broadcast to all users:"

func BroadcastConfirmText(text string, total int) string {
	return fmt.Sprintf("📢 Send this to %d user(s)?\n\n—\n%s\n—", total, text)
}

func BroadcastResultText(sent, failed int) string {
	if failed == 0 {
		return fmt.Sprintf("✅ Sent to %d user(s).", sent)
	}
	return fmt.Sprintf("✅ Sent to %d user(s). %d failed (likely blocked the bot).", sent, failed)
}

func valueOrDash(s *string) string {
	if s == nil || *s == "" {
		return "—"
	}
	return *s
}

func activeLabel(active bool) string {
	if active {
		return "active"
	}
	return "not active"
}
