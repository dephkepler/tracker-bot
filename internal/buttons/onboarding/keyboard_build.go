package onboarding

import (
	"fmt"
	challengebtn "tracker-bot/internal/buttons/challenge"
	learningbtn "tracker-bot/internal/buttons/learning"
	"tracker-bot/pkg/buttonbuilder"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StepInlineMenu renders one tour step's navigation: Back (if not the first
// step), Next/Finish, and Skip — except the last step, which instead offers
// direct jumps into each real feature plus Home.
func StepInlineMenu(step int) tgbotapi.InlineKeyboardMarkup {
	if step >= StepCount-1 {
		return buttonbuilder.IK(
			buttonbuilder.IR(buttonbuilder.IB(ButtonGoTrack, "back_to_main")), // literal used by every Track inline menu's own Back
			buttonbuilder.IR(buttonbuilder.IB(ButtonGoLearning, learningbtn.LearningCBBackMain)),
			buttonbuilder.IR(buttonbuilder.IB(ButtonGoChallenges, challengebtn.CBBackList)),
			buttonbuilder.IR(buttonbuilder.IB(ButtonHome, "go_home")),
		)
	}

	nav := make([]tgbotapi.InlineKeyboardButton, 0, 3)
	if step > 0 {
		nav = append(nav, buttonbuilder.IB(ButtonBack, fmt.Sprintf("%s%d", CBGoto, step-1)))
	}
	nav = append(nav, buttonbuilder.IB(ButtonNext, fmt.Sprintf("%s%d", CBGoto, step+1)))

	return buttonbuilder.IK(
		nav,
		buttonbuilder.IR(buttonbuilder.IB(ButtonSkip, CBSkip)),
	)
}
