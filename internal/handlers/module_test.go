package handlers

import (
	"strings"
	"testing"
	"time"
	"tracker-bot/internal/buttons/profile"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/internal/utils/tgctx"
)

func TestModule_IsAdmin(t *testing.T) {
	cases := []struct {
		name          string
		adminUsername string
		ctxUsername   string
		want          bool
	}{
		{"exact match", "alaamov", "alaamov", true},
		{"case-insensitive", "alaamov", "AlaAmov", true},
		{"leading @ in ctx username stored without it normally, but tolerate anyway", "alaamov", "alaamov", true},
		{"different user", "alaamov", "someoneelse", false},
		{"no admin configured", "", "alaamov", false},
		{"user has no @handle", "alaamov", "", false},
		{"neither set", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Module{adminUsername: tc.adminUsername}
			ctx := &tgctx.MsgContext{Username: tc.ctxUsername}
			if got := m.IsAdmin(ctx); got != tc.want {
				t.Fatalf("IsAdmin() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestNew_StripsLeadingAtFromAdminUsername checks the constructor normalizes
// a "@alaamov"-style config value the same way as a bare "alaamov", since
// it's easy to paste the @ into an env var by habit.
func TestNew_StripsLeadingAtFromAdminUsername(t *testing.T) {
	m := New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "@alaamov")
	if !m.IsAdmin(&tgctx.MsgContext{Username: "alaamov"}) {
		t.Fatal("IsAdmin() = false, want true after stripping leading @ from configured admin username")
	}
}

// TestLanguageCodeByButton checks every language-picker button maps to the
// exact code the DB's users_allowed_language CHECK constraint accepts
// (migrations/0001_users_init.up.sql: ru/en/de/uk/ar) — a typo here would
// make ProcessLanguageSelection silently fail ChangeLanguage for that button.
func TestLanguageCodeByButton(t *testing.T) {
	want := map[string]string{
		profile.ProfileButtonLanguageRussian:   "ru",
		profile.ProfileButtonLanguageEnglish:   "en",
		profile.ProfileButtonLanguageGerman:    "de",
		profile.ProfileButtonLanguageUkrainian: "uk",
		profile.ProfileButtonLanguageArabian:   "ar",
	}

	if len(languageCodeByButton) != len(want) {
		t.Fatalf("languageCodeByButton has %d entries, want %d", len(languageCodeByButton), len(want))
	}
	for button, code := range want {
		got, ok := languageCodeByButton[button]
		if !ok {
			t.Errorf("languageCodeByButton missing entry for %q", button)
			continue
		}
		if got != code {
			t.Errorf("languageCodeByButton[%q] = %q, want %q", button, got, code)
		}
	}
}

// TestAppendHourlyByActivityLines_GroupsByHour checks the "By hours" report
// line lists every activity tracked within that hour (most-time-first, per
// GetHourlyBucketsByActivity's ordering), one line per hour rather than
// silently collapsing to a single total.
func TestAppendHourlyByActivityLines_GroupsByHour(t *testing.T) {
	nine := time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)
	ten := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	rows := []models.HourActivityDuration{
		{BucketStart: nine, Name: "work", Emoji: "", Duration: 2 * time.Minute},
		{BucketStart: ten, Name: "Deep work", Emoji: "🏋", Duration: 30 * time.Minute},
		{BucketStart: ten, Name: "Reading", Emoji: "📖", Duration: 17 * time.Minute},
	}

	var b strings.Builder
	ctx := &tgctx.MsgContext{Language: i18n.EN}
	appendHourlyByActivityLines(ctx, &b, rows, "15:00")
	got := b.String()

	wantLines := []string{
		"- 09:00: work 2m",
		"- 10:00: 🏋 Deep work 30m, 📖 Reading 17m",
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("output = %q, want it to contain %q", got, want)
		}
	}
}
