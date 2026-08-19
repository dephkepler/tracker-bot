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
