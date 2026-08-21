package textbuilder

import "strings"

func strOrDash(s *string) string {
	if s == nil || *s == "" {
		return "—"
	}
	return *s
}

func escapeMarkdown(s string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
	)
	return r.Replace(s)
}

func StrOrDashMD(s *string) string {
	return escapeMarkdown(strOrDash(s))
}
