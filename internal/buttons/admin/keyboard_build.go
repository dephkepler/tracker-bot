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

const (
	LabelPrev = "◀ Prev"
	LabelNext = "Next ▶"
	LabelHome = "🏠 Home"
)

// Reply buttons. ReplyButtonBack/ReplyButtonHome intentionally reuse the
// exact same text as track.TrackButtonBack/TrackButtonBackHome so the
// existing dispatcher cases for those two buttons keep handling them —
// screen-specific behavior is added there by extending the isScreen switch,
// not by duplicating a case for an identical button string.
const (
	ReplyButtonUsers = "👥 Users"
	ReplyButtonBack  = "◀ Back"
	ReplyButtonHome  = "🏠 Home"
)

// MenuReplyMenu is the admin landing screen: pick "Users" to see the
// listing, or leave via Back/Home.
func MenuReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonUsers)),
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

// UsersReplyMenu is shown while browsing the paginated users list — Back
// returns to the admin landing screen (not Home).
func UsersReplyMenu() tgbotapi.ReplyKeyboardMarkup {
	return buttonbuilder.RK(
		buttonbuilder.RR(buttonbuilder.RB(ReplyButtonBack), buttonbuilder.RB(ReplyButtonHome)),
	)
}

// UsersInlineMenu renders one row per user (display-only, "noop" — this is
// a listing, not a picker) plus Prev/Next paging and a way back home.
func UsersInlineMenu(users []models.AdminUserRow, offset, limit, total int) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(users)+2)
	for _, u := range users {
		label := fmt.Sprintf("#%d %s — %s", u.DBID, usernameLabel(u.UserName), u.CreatedAt.Format("2006-01-02"))
		rows = append(rows, buttonbuilder.IR(buttonbuilder.IB(label, "noop")))
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
