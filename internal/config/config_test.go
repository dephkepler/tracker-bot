package config

import (
	"strings"
	"testing"
	"time"
)

func validWeb() Web {
	return Web{
		Enabled:          true,
		Addr:             ":8090",
		PublicOrigin:     "https://132-243-194-137.sslip.io:8443",
		BotUsername:      "trackerbot",
		MiniAppShortName: "dashboard",
		InitDataMaxAge:   24 * time.Hour,
		MaxInflight:      8,
	}
}

func TestWebValidateAcceptsAWorkingConfig(t *testing.T) {
	if err := validWeb().Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

// The dev bypass pins every request to one identity. Reaching it from a public
// origin would serve one person's data to anyone who found the URL, so this is
// a refusal to boot rather than a warning.
func TestWebValidateRefusesDevBypassOnPublicOrigin(t *testing.T) {
	w := validWeb()
	w.DevTgUserID = 424242

	err := w.Validate()
	if err == nil {
		t.Fatal("Validate accepted a dev bypass alongside an https origin")
	}
	if !strings.Contains(err.Error(), "WEB_DEV_TG_USER_ID") {
		t.Fatalf("error should name the offending setting, got: %v", err)
	}
}

func TestWebValidateAllowsDevBypassLocally(t *testing.T) {
	w := validWeb()
	w.DevTgUserID = 424242
	w.PublicOrigin = "http://localhost:3000"

	if err := w.Validate(); err != nil {
		t.Fatalf("Validate on a local origin: %v", err)
	}
}

func TestWebValidateRejectsBadValues(t *testing.T) {
	cases := map[string]func(*Web){
		"empty addr":         func(w *Web) { w.Addr = "" },
		"addr not host:port": func(w *Web) { w.Addr = "8090" },
		"zero inflight":      func(w *Web) { w.MaxInflight = 0 },
		"negative inflight":  func(w *Web) { w.MaxInflight = -1 },
		"zero max age":       func(w *Web) { w.InitDataMaxAge = 0 },
		"origin not a url":   func(w *Web) { w.PublicOrigin = "not a url" },
	}

	for name, break_ := range cases {
		t.Run(name, func(t *testing.T) {
			w := validWeb()
			break_(&w)
			if err := w.Validate(); err == nil {
				t.Fatal("Validate accepted it")
			}
		})
	}
}

// A disabled dashboard is never validated, so an existing .env that knows
// nothing about these settings must still start.
func TestValidateSkipsADisabledDashboard(t *testing.T) {
	cfg := Config{Web: Web{Enabled: false}} // every other field is a zero value
	if err := cfg.validate(); err != nil {
		t.Fatalf("a disabled dashboard must not be validated, got: %v", err)
	}
}

func TestMiniAppURL(t *testing.T) {
	cases := []struct {
		name     string
		username string
		short    string
		want     string
	}{
		{"plain", "trackerbot", "dashboard", "https://t.me/trackerbot/dashboard"},
		// Pasting the @ into an env var is an easy habit, same reason
		// handlers.New strips it from the admin username.
		{"leading at", "@trackerbot", "dashboard", "https://t.me/trackerbot/dashboard"},
		// No username means the Mini App is not registered yet, and the profile
		// menu must not draw a button that goes nowhere.
		{"no username", "", "dashboard", ""},
		{"no short name", "trackerbot", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Web: Web{BotUsername: tc.username, MiniAppShortName: tc.short}}
			if got := cfg.MiniAppURL(); got != tc.want {
				t.Fatalf("MiniAppURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDashboardURL(t *testing.T) {
	cases := map[string]string{
		"https://host:8443":  "https://host:8443/",
		"https://host:8443/": "https://host:8443/",
		"":                   "",
	}
	for origin, want := range cases {
		w := Web{PublicOrigin: origin}
		if got := w.DashboardURL(); got != want {
			t.Fatalf("DashboardURL(%q) = %q, want %q", origin, got, want)
		}
	}
}
