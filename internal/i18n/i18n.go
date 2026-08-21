// Package i18n exists because reply-keyboard dispatch matches literal button text, which breaks once labels are per-language.
package i18n

import (
	"fmt"
	"maps"
	"strings"
)

// Lang values must stay in sync with the DB's users_allowed_language CHECK constraint.
type Lang string

const (
	RU Lang = "ru"
	EN Lang = "en"
	DE Lang = "de"
	UK Lang = "uk"
	AR Lang = "ar"

	Default = EN
)

var All = []Lang{RU, EN, DE, UK, AR}

// Normalize never errors; an empty/unknown code silently falls back to Default.
func Normalize(code string) Lang {
	switch Lang(strings.ToLower(strings.TrimSpace(code))) {
	case RU, EN, DE, UK, AR:
		return Lang(strings.ToLower(strings.TrimSpace(code)))
	default:
		return Default
	}
}

// T falls back to Default lang, then to the bare key, so a missing translation is visible, not blank.
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

type entry map[Lang]string

// merges every catalog_*.go map; safe because Go initializes package vars in dependency order, not file order.
var catalog = mergeCatalogs(
	catalogCommon,
	catalogEntry,
	catalogProfile,
	catalogTrack,
	catalogTrackReports,
	catalogLearning,
	catalogRoadmap,
	catalogChallenge,
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

// reverse only holds short stable button text, not templates with fmt args — see Key.
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

// Key falls back to Default for not-yet-translated buttons; ok is false for dynamic text like timer minutes.
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
