package profile

import (
	"testing"

	"tracker-bot/internal/i18n"
)

// The dashboard row appears only once a Mini App URL exists. Before BotFather
// registration there is nothing to open, and a button that goes nowhere is
// worse than no button — so the empty case is the one worth pinning.
func TestProfileEntryInlineMenuDashboardRow(t *testing.T) {
	const url = "https://t.me/stepbystep_trackerbot/dashboard"

	withURL := ProfileEntryInlineMenu(i18n.RU, false, url)
	withoutURL := ProfileEntryInlineMenu(i18n.RU, false, "")

	if len(withURL.InlineKeyboard) != len(withoutURL.InlineKeyboard)+1 {
		t.Fatalf("rows with url = %d, without = %d; want exactly one more",
			len(withURL.InlineKeyboard), len(withoutURL.InlineKeyboard))
	}

	last := withURL.InlineKeyboard[len(withURL.InlineKeyboard)-1]
	if len(last) != 1 {
		t.Fatalf("dashboard row has %d buttons, want 1", len(last))
	}
	btn := last[0]
	if btn.URL == nil {
		t.Fatal("dashboard button carries no URL — a callback button would need dispatcher routing")
	}
	if *btn.URL != url {
		t.Fatalf("URL = %q, want %q", *btn.URL, url)
	}
	// A URL button must not also carry callback data: Telegram would ignore the
	// link and the tap would fall through to the dispatcher as an unknown
	// callback.
	if btn.CallbackData != nil {
		t.Fatalf("dashboard button also carries callback data %q", *btn.CallbackData)
	}
	if btn.Text != i18n.T(i18n.RU, i18n.KeyProfileButtonDashboard) {
		t.Fatalf("label = %q, want the translated dashboard label", btn.Text)
	}
}

// Every supported language must render a label, or some users get a bare key.
func TestDashboardButtonIsTranslatedEverywhere(t *testing.T) {
	for _, lang := range i18n.All {
		label := i18n.T(lang, i18n.KeyProfileButtonDashboard)
		if label == i18n.KeyProfileButtonDashboard {
			t.Fatalf("%s: label falls back to the raw key", lang)
		}
	}
}
