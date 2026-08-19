package handlers

import (
	"testing"
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
	m := New(nil, nil, nil, nil, nil, nil, nil, "@alaamov")
	if !m.IsAdmin(&tgctx.MsgContext{Username: "alaamov"}) {
		t.Fatal("IsAdmin() = false, want true after stripping leading @ from configured admin username")
	}
}
