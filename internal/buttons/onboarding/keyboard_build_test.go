package onboarding

import (
	"fmt"
	"testing"
)

func TestStepInlineMenu_FirstStepHasNoBack(t *testing.T) {
	menu := StepInlineMenu(0)
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			if btn.Text == ButtonBack {
				t.Fatal("step 0 must not offer a Back button")
			}
		}
	}
}

func TestStepInlineMenu_MiddleStepHasBackAndNext(t *testing.T) {
	menu := StepInlineMenu(1)
	var hasBack, hasNext bool
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			switch btn.Text {
			case ButtonBack:
				hasBack = true
				if btn.CallbackData == nil || *btn.CallbackData != fmt.Sprintf("%s%d", CBGoto, 0) {
					t.Errorf("Back callback = %v, want %s0", btn.CallbackData, CBGoto)
				}
			case ButtonNext:
				hasNext = true
				if btn.CallbackData == nil || *btn.CallbackData != fmt.Sprintf("%s%d", CBGoto, 2) {
					t.Errorf("Next callback = %v, want %s2", btn.CallbackData, CBGoto)
				}
			}
		}
	}
	if !hasBack || !hasNext {
		t.Fatalf("middle step: hasBack=%v hasNext=%v, want both true", hasBack, hasNext)
	}
}

// TestStepInlineMenu_LastStepOffersJumpsNotNext checks the final step
// swaps navigation for direct jumps into each real feature — no dead-end.
func TestStepInlineMenu_LastStepOffersJumpsNotNext(t *testing.T) {
	menu := StepInlineMenu(StepCount - 1)
	var hasNext, hasTrack, hasLearning, hasChallenges, hasHome bool
	for _, row := range menu.InlineKeyboard {
		for _, btn := range row {
			switch btn.Text {
			case ButtonNext:
				hasNext = true
			case ButtonGoTrack:
				hasTrack = true
			case ButtonGoLearning:
				hasLearning = true
			case ButtonGoChallenges:
				hasChallenges = true
			case ButtonHome:
				hasHome = true
			}
		}
	}
	if hasNext {
		t.Error("last step must not offer Next (nothing after it)")
	}
	if !hasTrack || !hasLearning || !hasChallenges || !hasHome {
		t.Fatalf("last step jumps: track=%v learning=%v challenges=%v home=%v, want all true", hasTrack, hasLearning, hasChallenges, hasHome)
	}
}

func TestStepText_ClampsOutOfRangeIndices(t *testing.T) {
	if StepText(-5) != StepText(0) {
		t.Error("negative step should clamp to step 0")
	}
	if StepText(999) != StepText(StepCount-1) {
		t.Error("out-of-range step should clamp to the last step")
	}
}
