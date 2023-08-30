package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-micro.dev/v4/logger"
	"sync"
)

type Bot struct {
	l   logger.Logger
	wg  sync.WaitGroup
	api *tgbotapi.BotAPI
	h   MessageHandler
}

func New(token string, handler MessageHandler) (*Bot, error) {
	var err error
	bot := &Bot{
		l: logger.DefaultLogger.Fields(map[string]interface{}{"from": "bot"}),
		h: handler,
	}

	bot.api, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.wg.Add(1)
	go func() {
		defer bot.wg.Done()
		bot.loop()
	}()

	return bot, nil
}

func (bot *Bot) loop() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(updateConfig)

	if err != nil {
		panic(fmt.Sprintf("Get updates failed: %s", err))
	}

	for {
		select {
		case update := <-updates:
			var message *tgbotapi.Message
			if update.Message == nil {
				if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
					message = update.CallbackQuery.Message
					message.Text = update.CallbackQuery.Data
					message.From = update.CallbackQuery.From
				}
			} else {
				message = update.Message
				if update.Message.From != nil {
					incomingMessagesCounter.WithLabelValues(update.Message.From.UserName).Inc()
				}
			}
			if message == nil {
				continue
			}

			bot.l.Logf(logger.InfoLevel, "Got message: '%s' from @%s", message.Text, message.From.UserName)
			bot.wg.Add(1)
			go bot.handleMessage(message)
		}
	}
}

func (bot *Bot) handleMessage(message *tgbotapi.Message) {
	defer bot.wg.Done()
	responses := bot.h.HandleMessage(message.MessageID, message.From.ID, message.From.UserName, message.Text)
	for _, response := range responses {
		msg := response.Compose(message.Chat.ID)
		if _, err := bot.api.Send(msg); err != nil {
			bot.l.Logf(logger.ErrorLevel, "Send message to %s failed: %s", message.From.UserName, err)
		}
	}
}

func (bot *Bot) Stop() {

}
