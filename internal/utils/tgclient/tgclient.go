package tgclient

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	StopReceivingUpdates()
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

func New(token string) (*tgbotapi.BotAPI, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram token is empty")
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return api, nil
}
