package admin

import (
	"testing"
	"tracker-bot/internal/models"
)

func TestUsersInlineMenu_Paging(t *testing.T) {
	users := []models.AdminUserRow{{DBID: 1}}

	cases := []struct {
		name          string
		offset, limit int
		total         int
		wantPrev      bool
		wantNext      bool
	}{
		{"first page, more pages after", 0, 15, 40, false, true},
		{"middle page", 15, 15, 40, true, true},
		{"last page, exact fit", 30, 15, 45, true, false},
		{"only page", 0, 15, 1, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			menu := UsersInlineMenu(users, tc.offset, tc.limit, tc.total)

			var hasPrev, hasNext bool
			for _, row := range menu.InlineKeyboard {
				for _, btn := range row {
					switch btn.Text {
					case LabelPrev:
						hasPrev = true
					case LabelNext:
						hasNext = true
					}
				}
			}

			if hasPrev != tc.wantPrev {
				t.Errorf("prev button present = %v, want %v", hasPrev, tc.wantPrev)
			}
			if hasNext != tc.wantNext {
				t.Errorf("next button present = %v, want %v", hasNext, tc.wantNext)
			}
		})
	}
}

func TestUsersInlineMenu_AlwaysHasHomeButton(t *testing.T) {
	menu := UsersInlineMenu(nil, 0, 15, 0)
	found := false
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			if btn.Text == LabelHome {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("UsersInlineMenu with no users still needs a way back home")
	}
}

func TestUsersInlineMenu_RowsAreClickable(t *testing.T) {
	users := []models.AdminUserRow{{DBID: 7}}
	menu := UsersInlineMenu(users, 0, 15, 1)

	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == nil {
				continue
			}
			if *btn.CallbackData == "noop" {
				t.Fatalf("user row callback = %q, want a real admin:user:detail: callback, not noop", *btn.CallbackData)
			}
		}
	}
}

// TestUserDetailInlineMenu_HidesDeleteForSelf guards against an admin
// locking themselves out by deleting their own account.
func TestUserDetailInlineMenu_HidesDeleteForSelf(t *testing.T) {
	menu := UserDetailInlineMenu(1, true)
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			if btn.Text == LabelDelete {
				t.Fatal("UserDetailInlineMenu(isSelf=true) must not offer a delete button")
			}
		}
	}
}

func TestUserDetailInlineMenu_OffersDeleteForOthers(t *testing.T) {
	menu := UserDetailInlineMenu(1, false)
	found := false
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			if btn.Text == LabelDelete {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("UserDetailInlineMenu(isSelf=false) must offer a delete button")
	}
}

func TestBroadcastConfirmInlineMenu_HasSendAndCancel(t *testing.T) {
	menu := BroadcastConfirmInlineMenu()
	var hasSend, hasCancel bool
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			switch btn.Text {
			case LabelSend:
				hasSend = true
			case LabelCancel:
				hasCancel = true
			}
		}
	}
	if !hasSend || !hasCancel {
		t.Fatalf("BroadcastConfirmInlineMenu: hasSend=%v hasCancel=%v, want both true", hasSend, hasCancel)
	}
}
