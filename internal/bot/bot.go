package bot

import (
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-micro.dev/v4/logger"
	"sync"
)

type Bot struct {
	l   logger.Logger
	api *tgbotapi.BotAPI
	wg  sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	incoming chan *messaging.Incoming
	outgoing chan *messaging.Outgoing
}

const maxMessages = 10000

func New(token string) (*Bot, error) {
	var err error
	bot := &Bot{
		l:        logger.DefaultLogger.Fields(map[string]interface{}{"from": "bot"}),
		incoming: make(chan *messaging.Incoming, maxMessages),
		outgoing: make(chan *messaging.Outgoing, maxMessages),
	}

	bot.api, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.ctx, bot.cancel = context.WithCancel(context.Background())

	bot.wg.Add(1)
	go func() {
		defer bot.wg.Done()
		bot.loop()
	}()

	return bot, nil
}

func (bot *Bot) Incoming() <-chan *messaging.Incoming {
	return bot.incoming
}

func (bot *Bot) Outgoing() chan<- *messaging.Outgoing {
	return bot.outgoing
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

			msg := messaging.Incoming{
				ID:       message.MessageID,
				ChatID:   message.Chat.ID,
				UserID:   message.From.ID,
				UserName: message.From.UserName,
				Text:     message.Text,
			}
			bot.l.Logf(logger.DebugLevel, "Message '%s' from @%s [ %d ]", message.Text, message.From.UserName, message.From.ID)
			bot.incoming <- &msg

		case message := <-bot.outgoing:
			if _, err = bot.api.Send(message.Message.Compose(message.ChatID)); err != nil {
				bot.l.Logf(logger.ErrorLevel, "Send message failed: %s")
			}
		case <-bot.ctx.Done():
			return
		}
	}
}

func (bot *Bot) Stop() {
	bot.cancel()
	bot.wg.Wait()

	close(bot.incoming)
	close(bot.outgoing)
}
