package i18n

import "testing"

// TestCatalog_AllLanguagesPresent guards against a phase adding a key with
// some languages missing — T() would silently fall back to English for
// that language/key forever otherwise, easy to miss in review.
func TestCatalog_AllLanguagesPresent(t *testing.T) {
	for key, e := range catalog {
		for _, l := range All {
			if _, ok := e[l]; !ok {
				t.Errorf("key %q missing translation for %q", key, l)
			}
		}
	}
}

// TestCatalog_NoTextCollisionsWithinLanguage guards the mechanism Key()
// depends on: if two different keys render to the same text in the same
// language, the reverse lookup is ambiguous and dispatch could resolve a
// button tap to the wrong action.
func TestCatalog_NoTextCollisionsWithinLanguage(t *testing.T) {
	seen := make(map[Lang]map[string]string)
	for key, e := range catalog {
		for lang, text := range e {
			if seen[lang] == nil {
				seen[lang] = make(map[string]string)
			}
			if other, ok := seen[lang][text]; ok && other != key {
				t.Errorf("lang %q: text %q is used by both key %q and key %q", lang, text, other, key)
				continue
			}
			seen[lang][text] = key
		}
	}
}

func TestT_FormatsWithArgs(t *testing.T) {
	got := T(EN, KeyProfileTimezoneSaved, "Europe/Berlin")
	want := "✅ Time zone set to Europe/Berlin"
	if got != want {
		t.Fatalf("T() = %q, want %q", got, want)
	}
}

func TestT_FallsBackToDefaultLanguage(t *testing.T) {
	// Every phase-1 key currently has all 5 languages (enforced by
	// TestCatalog_AllLanguagesPresent), so exercise the fallback path
	// directly against lookup instead of relying on an incomplete key.
	if _, ok := lookup(Lang("fr"), KeyCommonHome); !ok {
		t.Fatal("lookup with an unsupported language should still fall back to Default, not fail")
	}
	got := T(Lang("fr"), KeyCommonHome)
	want := T(Default, KeyCommonHome)
	if got != want {
		t.Fatalf("T() with unsupported lang = %q, want fallback to Default %q", got, want)
	}
}

func TestT_MissingKeyReturnsKeyItself(t *testing.T) {
	const missing = "does.not.exist"
	if got := T(EN, missing); got != missing {
		t.Fatalf("T() for missing key = %q, want the key itself %q (visible failure, not blank text)", got, missing)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]Lang{
		"ru": RU, "RU": RU, " ru ": RU,
		"en": EN, "": EN, "fr": EN, "xx": EN,
		"de": DE, "uk": UK, "ar": AR,
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestKey_RoundTripsWithT(t *testing.T) {
	for _, lang := range All {
		for _, key := range []string{
			KeyCommonBack, KeyCommonHome, KeyCommonCancel, KeyCommonAdmin,
			KeyEntryButtonProfile, KeyEntryButtonTrack, KeyEntryButtonLearning, KeyEntryButtonSubscription,
		} {
			text := T(lang, key)
			gotKey, ok := Key(lang, text)
			if !ok {
				t.Errorf("Key(%q, %q): ok = false, want true", lang, text)
				continue
			}
			if gotKey != key {
				t.Errorf("Key(%q, %q) = %q, want %q", lang, text, gotKey, key)
			}
		}
	}
}

func TestKey_UnknownTextIsNotFound(t *testing.T) {
	if _, ok := Key(EN, "this is not a translated button"); ok {
		t.Fatal("Key() for arbitrary text: ok = true, want false")
	}
}

// TestKey_FallsBackToDefaultAcrossPhases guards the mechanism that makes a
// phased rollout safe: a screen not yet converted to i18n always renders
// English text (the literal Go constants it always had), regardless of the
// viewer's own language. A non-English user tapping that still-English
// button must still resolve correctly, or navigation breaks for them on
// every screen this phase hasn't reached yet.
func TestKey_FallsBackToDefaultAcrossPhases(t *testing.T) {
	englishText := T(EN, KeyCommonHome)
	for _, lang := range All {
		if lang == EN {
			continue
		}
		key, ok := Key(lang, englishText)
		if !ok {
			t.Errorf("Key(%q, %q) [unconverted-screen English text]: ok = false, want true", lang, englishText)
			continue
		}
		if key != KeyCommonHome {
			t.Errorf("Key(%q, %q) = %q, want %q", lang, englishText, key, KeyCommonHome)
		}
	}
}
