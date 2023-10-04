package artist

import (
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/music"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
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
	Attributes: command.Attributes{
		AuthRequired: true,
		CanRepeat:    true,
	},
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	c := artistCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "artist"}),
	}
	c.f = formatter.New(c.l, c.interlayer.Registry)
	return &c
}

func (c artistCommand) Do(ctx command.Context) []messaging.ChatMessage {
	if len(ctx.Arguments) == 0 {
		return messaging.NewSingleMessage("Имя исполнителя?", ctx.ReplyID)
	}

	cli, auth := c.interlayer.Discovery.New(ctx.Token)

	reqCtx, cancel := context.WithTimeout(ctx.Ctx, searchTimeout)
	defer cancel()

	req := music.SearchMusicParams{
		Limit:   utils.ToPointer(maxResult),
		Q:       ctx.Arguments.String(),
		Type:    utils.ToPointer("artist"),
		Context: reqCtx,
	}

	resp, err := cli.Music.SearchMusic(&req, auth)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Search music failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, ctx.ReplyID)
	}
	if len(resp.Payload.Results) == 0 {
		return messaging.NewSingleMessage(command.NothingFound, ctx.ReplyID)
	}

	messages := make([]messaging.ChatMessage, 0, len(resp.Payload.Results))
	for _, item := range resp.Payload.Results {
		msg := c.f.FormatSearchMusicResult(item, ctx.ReplyID)
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	c.l.Logf(logger.InfoLevel, "Got %d results", len(messages))
	return messaging.Reverse(messages)
}
