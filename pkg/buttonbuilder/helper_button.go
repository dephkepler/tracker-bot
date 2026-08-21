package buttonbuilder

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func IB(text, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, data)
}

func IR(btns ...tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(btns...)
}

func IK(rows ...[]tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func RB(text string) tgbotapi.KeyboardButton {
	return tgbotapi.NewKeyboardButton(text)
}

// sets RequestLocation so Telegram prompts the user to share their location when tapped
func RBLocation(text string) tgbotapi.KeyboardButton {
	btn := tgbotapi.NewKeyboardButton(text)
	btn.RequestLocation = true
	return btn
}

func RR(btns ...tgbotapi.KeyboardButton) []tgbotapi.KeyboardButton {
	return tgbotapi.NewKeyboardButtonRow(btns...)
}

func RK(rows ...[]tgbotapi.KeyboardButton) tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	return kb
}
