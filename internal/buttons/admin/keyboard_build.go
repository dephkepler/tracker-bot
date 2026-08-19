// Package admin renders the admin-only "who's using the bot" screen: a
// user count plus a paged, read-only inline listing.
package admin

import (
	"fmt"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CBOpen opens the admin screen (page 0) — used by entry points that live
// outside the entry-screen reply keyboard, e.g. the Profile screen's inline
// menu (see internal/buttons/profile.ProfileEntryInlineMenu).
const CBOpen = "admin:open"

// CBUsersPage pages through the user listing: callback data is
// "<CBUsersPage><offset>".
const CBUsersPage = "admin:users:page:"

// CBUserDetail opens one user's drill-down view: "<CBUserDetail><dbID>".
const CBUserDetail = "admin:user:detail:"

// CBUserDeleteAsk shows the delete confirmation step for one user.
// CBUserDeleteConfirm actually performs the deletion.
const (
	CBUserDeleteAsk     = "admin:user:delete:ask:"
	CBUserDeleteConfirm = "admin:user:delete:confirm:"
)

// CBOverview refreshes/opens the bot-wide overview stats.
const CBOverview = "admin:overview:open"

// CBBroadcastConfirm/CBBroadcastCancel act on the pending broadcast text
// held in the dispatcher session (see userSession.pendingBroadcastText).
const (
	CBBroadcastConfirm = "admin:broadcast:confirm"
	CBBroadcastCancel  = "admin:broadcast:cancel"
)

const (
	LabelPrev = "◀ Prev"
	LabelNext = "Next ▶"
	LabelHome = "🏠 Home"
	LabelBack = "◀ Back"

	LabelDelete        = "🗑 Delete user"
	LabelConfirmDelete = "⚠️ Confirm delete"
	LabelCancel        = "❌ Cancel"

	LabelSend = "✅ Send"
)

// Reply buttons. ReplyButtonBack/ReplyButtonHome intentionally reuse the
// exact same text as track.TrackButtonBack/TrackButtonBackHome so the
// existing dispatcher cases for those two buttons keep handling them —
// screen-specific behavior is added there by extending the isScreen switch,
// not by duplicating a case for an identical button string.
const (
	ReplyButtonUsers     = "👥 Users"
	ReplyButtonBroadcast = "📢 Broadcast"
	ReplyButtonOverview  = "📊 Overview"
	ReplyButtonBack      = "◀ Back"
	ReplyButtonHome      = "🏠 Home"
)

// MenuReplyMenu is the admin landing screen: pick "Users" to see the
// listing, "Broadcast" to message everyone, "Overview" for bot-wide stats,
// or leave via Back/Home.
func MenuReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonUsers)),
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBroadcast), buttonbuilder.RB(ReplyButtonOverview)),
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

// BroadcastWaitingReplyMenu is shown while the bot is waiting for the
// broadcast message text.
func BroadcastWaitingReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LabelCancel)),
	)
}

// UsersReplyMenu is shown while browsing the paginated users list — Back
// returns to the admin landing screen (not Home).
func UsersReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

// UsersInlineMenu renders one row per user — tap opens that user's
// drill-down view (see UserDetailInlineMenu) — plus Prev/Next paging and a
// way back home.
func UsersInlineMenu(users []models.AdminUserRow, offset, limit, total int) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(users)+2)
	for _, u := range users {
		label := fmt.Sprintf("#%d %s — %s", u.DBID, usernameLabel(u.UserName), u.CreatedAt.Format("2006-01-02"))
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, fmt.Sprintf("%s%d", CBUserDetail, u.DBID))))
	}

	nav := make([]tgbotapi.InlineKeyboardButton, 0, 2)
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		nav = append(nav, buttonbuilder.IB(LabelPrev, fmt.Sprintf("%s%d", CBUsersPage, prevOffset)))
	}
	if offset+limit < total {
		nav = append(nav, buttonbuilder.IB(LabelNext, fmt.Sprintf("%s%d", CBUsersPage, offset+limit)))
	}
	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LabelHome, "go_home")))
	return buttonbuilder.IK(rows...)
}

func usernameLabel(userName *string) string {
	if userName == nil || *userName == "" {
		return "(no username)"
	}
	return "@" + *userName
}

// UserDetailInlineMenu shows a delete action (omitted entirely when viewing
// the admin's own account — see isSelf) plus a way back to the list.
func UserDetailInlineMenu(dbID int64, isSelf bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 2)
	if !isSelf {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LabelDelete, fmt.Sprintf("%s%d", CBUserDeleteAsk, dbID))))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LabelBack, CBUsersPage+"0")))
	return buttonbuilder.IK(rows...)
}

// BackToUsersInlineMenu is a single "back to the users list" row, used
// after an action that leaves nothing else useful to show (e.g. a
// completed delete).
func BackToUsersInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(LabelBack, CBUsersPage+"0")),
	)
}

// UserDeleteConfirmInlineMenu is the "are you sure" step before an
// irreversible delete.
func UserDeleteConfirmInlineMenu(dbID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LabelConfirmDelete, fmt.Sprintf("%s%d", CBUserDeleteConfirm, dbID)),
			buttonbuilder.IB(LabelCancel, fmt.Sprintf("%s%d", CBUserDetail, dbID)),
		),
	)
}

// BroadcastConfirmInlineMenu is the "are you sure" step before sending a
// message to every registered user.
func BroadcastConfirmInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LabelSend, CBBroadcastConfirm),
			buttonbuilder.IB(LabelCancel, CBBroadcastCancel),
		),
	)
}
