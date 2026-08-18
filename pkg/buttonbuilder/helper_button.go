// Package buttonbuilder provides small helpers to build Telegram keyboards.
package buttonbuilder

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// IB creates one inline button.
func IB(text, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, data)
}

// IR creates one inline keyboard row.
func IR(btns ...tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(btns...)
}

// IK creates an inline keyboard from rows.
func IK(rows ...[]tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// RB creates one reply button.
func RB(text string) tgbotapi.KeyboardButton {
	return tgbotapi.NewKeyboardButton(text)
}

// RBLocation creates a reply button that asks Telegram to prompt the user to
// share their current location when tapped.
func RBLocation(text string) tgbotapi.KeyboardButton {
	btn := tgbotapi.NewKeyboardButton(text)
	btn.RequestLocation = true
	return btn
}

// RR creates one reply keyboard row.
func RR(btns ...tgbotapi.KeyboardButton) []tgbotapi.KeyboardButton {
	return tgbotapi.NewKeyboardButtonRow(btns...)
}

// RK creates a resize-enabled reply keyboard from rows.
func RK(rows ...[]tgbotapi.KeyboardButton) tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	return kb
}
