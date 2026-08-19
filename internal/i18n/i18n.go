// Package i18n is the bot's translation catalog and lookup. It exists to
// solve a problem specific to Telegram reply keyboards: their buttons are
// dispatched by matching the literal text the user tapped (see
// internal/dispatcher/dispatcher.go's `switch ctx.Text { case ... }`
// blocks). Once button labels are rendered per-user-language, that literal
// match breaks — "◀ Back" and "◀ Zurück" are different strings for the same
// action. So every translatable string here is filed under a stable,
// language-independent key; T renders a key into display text for a given
// language, and Key does the reverse (rendered text -> key) so the
// dispatcher can switch on the stable key instead of a hardcoded literal.
package i18n

import (
	"fmt"
	"maps"
	"strings"
)

// Lang is a supported interface language code. Values match the DB's
// users_allowed_language CHECK constraint (migrations/0001_users_init.up.sql:
// language IN ('ru','en','de','uk','ar')) — the two are kept in sync
// deliberately, since a user's stored language is what selects their Lang.
type Lang string

const (
	RU Lang = "ru"
	EN Lang = "en"
	DE Lang = "de"
	UK Lang = "uk"
	AR Lang = "ar"

	// Default is used whenever a stored/requested language is empty or not
	// one of the five supported codes.
	Default = EN
)

// All lists every supported language, in the order they should generally be
// offered to a user (e.g. building a language picker).
var All = []Lang{RU, EN, DE, UK, AR}

// Normalize maps a raw language code (as stored on the user — possibly
// empty, mixed-case, or something unexpected) to a supported Lang, falling
// back to Default rather than erroring, since a bad/missing language must
// never block a user from using the bot.
func Normalize(code string) Lang {
	switch Lang(strings.ToLower(strings.TrimSpace(code))) {
	case RU, EN, DE, UK, AR:
		return Lang(strings.ToLower(strings.TrimSpace(code)))
	default:
		return Default
	}
}

// T renders key in lang, formatting with args via fmt.Sprintf when given.
// Falls back to the Default language if lang has no translation for key,
// and to the bare key itself if even that's missing — both are visible,
// greppable failure states in the chat rather than a blank message, so a
// missing translation is easy to notice and find.
func T(lang Lang, key string, args ...any) string {
	tpl, ok := lookup(lang, key)
	if !ok {
		return key
	}
	if len(args) == 0 {
		return tpl
	}
	return fmt.Sprintf(tpl, args...)
}

// entry is one translatable string: its template per language.
type entry map[Lang]string

// catalog is assembled from every catalog_*.go file's own map (one per UI
// area/phase — see catalog_common.go, catalog_entry.go, catalog_profile.go).
// Built via mergeCatalogs rather than per-file init() functions: Go
// initializes package-level vars in dependency order regardless of which
// file they're declared in, so this — and reverse below, which depends on
// catalog — are guaranteed to see every phase's entries by the time
// anything calls T/Key, however many catalog_*.go files exist.
var catalog = mergeCatalogs(
	catalogCommon,
	catalogEntry,
	catalogProfile,
	catalogTrack,
	catalogTrackReports,
)

func mergeCatalogs(parts ...map[string]entry) map[string]entry {
	merged := make(map[string]entry)
	for _, part := range parts {
		maps.Copy(merged, part)
	}
	return merged
}

func lookup(lang Lang, key string) (string, bool) {
	e, ok := catalog[key]
	if !ok {
		return "", false
	}
	if tpl, ok := e[lang]; ok {
		return tpl, true
	}
	if tpl, ok := e[Default]; ok {
		return tpl, true
	}
	return "", false
}

// reverse[lang][renderedText] = key, built once from catalog. Only meant
// for button-style keys (short, stable, no fmt args) — see Key.
var reverse = buildReverse()

func buildReverse() map[Lang]map[string]string {
	rev := make(map[Lang]map[string]string, len(All))
	for _, l := range All {
		rev[l] = make(map[string]string)
	}
	for key, e := range catalog {
		for lang, text := range e {
			rev[lang][text] = key
		}
	}
	return rev
}

// Key resolves a reply-button's rendered text back to its translation key,
// so the dispatcher can switch on the stable key instead of a
// language-specific literal. It tries lang first, then falls back to
// Default (English) — needed while the UI is only partially translated
// (this ships in phases, see internal/i18n's package doc): a screen that
// hasn't been converted yet always renders English regardless of the
// user's own language, and this fallback is what lets a Russian-language
// user's tap on that still-English button keep resolving correctly. Once
// every screen is converted this fallback simply never triggers for a
// non-English user's own presses, so it's harmless to leave in place.
//
// ok is false when text isn't a known translated string at all (e.g. it's
// not a button, or it's dynamic text like a timer's "N min" — those need
// their own prefix/suffix parsing, see
// internal/buttons/track.ParseTimerButtonMinutes).
func Key(lang Lang, text string) (string, bool) {
	if key, ok := reverse[lang][text]; ok {
		return key, true
	}
	if lang != Default {
		if key, ok := reverse[Default][text]; ok {
			return key, true
		}
	}
	return "", false
}
