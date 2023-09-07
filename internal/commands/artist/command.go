package artist

import (
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/music"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/formatter"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"go-micro.dev/v4/logger"
	"time"
)

type artistCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
	f          formatter.Formatter
}

const (
	maxResult     int64 = 3
	searchTimeout       = 30 * time.Second
)

var Command command.Type = command.Type{
	ID:      "artist",
	Title:   "Поиск исполнителей",
	Help:    "Позволяет добавить дискографию исполнителя целиком",
	Factory: New,
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	c := artistCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "artist"}),
	}
	c.f = formatter.New(c.l, c.interlayer.Registry)
	return &c
}

func (c artistCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	if len(arguments) == 0 {
		return messaging.NewSingleMessage("Имя исполнителя?", replyID)
	}
	var token = config.Config().Token // TODO: remove
	cli, auth := c.interlayer.Discovery.New(token)

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	req := music.SearchMusicParams{
		Limit:   utils.ToPointer(maxResult),
		Q:       arguments.String(),
		Type:    utils.ToPointer("artist"),
		Context: ctx,
	}

	resp, err := cli.Music.SearchMusic(&req, auth)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Search music failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	if len(resp.Payload.Results) == 0 {
		return messaging.NewSingleMessage(command.NothingFound, replyID)
	}

	messages := make([]messaging.ChatMessage, 0, len(resp.Payload.Results))
	for _, item := range resp.Payload.Results {
		msg := c.f.FormatSearchMusicResult(item, replyID)
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	c.l.Logf(logger.InfoLevel, "Got %d results", len(messages))
	return messages
}
