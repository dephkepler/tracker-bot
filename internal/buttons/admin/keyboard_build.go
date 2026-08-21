package admin

import (
	"fmt"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const CBOpen = "admin:open"

const CBUsersPage = "admin:users:page:"

const CBUserDetail = "admin:user:detail:"

const (
	CBUserDeleteAsk     = "admin:user:delete:ask:"
	CBUserDeleteConfirm = "admin:user:delete:confirm:"
)

const CBOverview = "admin:overview:open"

// Acts on the pending broadcast text held in userSession.pendingBroadcastText.
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

// ReplyButtonBack/Home intentionally match track.TrackButtonBack/BackHome's text so the dispatcher's existing case handles them too.
const (
	ReplyButtonUsers     = "👥 Users"
	ReplyButtonBroadcast = "📢 Broadcast"
	ReplyButtonOverview  = "📊 Overview"
	ReplyButtonBack      = "◀ Back"
	ReplyButtonHome      = "🏠 Home"
)

func MenuReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonUsers)),
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBroadcast), buttonbuilder.RB(ReplyButtonOverview)),
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

func BroadcastWaitingReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(LabelCancel)),
	)
}

func UsersReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

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

func UserDetailInlineMenu(dbID int64, isSelf bool) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 2)
	if !isSelf {
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LabelDelete, fmt.Sprintf("%s%d", CBUserDeleteAsk, dbID))))
	}
	rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(LabelBack, CBUsersPage+"0")))
	return buttonbuilder.IK(rows...)
}

func BackToUsersInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(buttonbuilder.IB(LabelBack, CBUsersPage+"0")),
	)
}

func UserDeleteConfirmInlineMenu(dbID int64) tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LabelConfirmDelete, fmt.Sprintf("%s%d", CBUserDeleteConfirm, dbID)),
			buttonbuilder.IB(LabelCancel, fmt.Sprintf("%s%d", CBUserDetail, dbID)),
		),
	)
}

func BroadcastConfirmInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return buttonbuilder.IK(
		buttonbuilder.IR(
			buttonbuilder.IB(LabelSend, CBBroadcastConfirm),
			buttonbuilder.IB(LabelCancel, CBBroadcastCancel),
		),
	)
}
