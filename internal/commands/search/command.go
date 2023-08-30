package search

import (
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/music"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"go-micro.dev/v4/logger"
	"time"
)

type searchCommand struct {
	f connectivity.Interlayer
	l logger.Logger
	r registry.Registry
}

const (
	maxResult     int64 = 4
	searchTimeout       = 30 * time.Second
)

var Command command.Type = command.Type{
	ID:      "search",
	Title:   "Поиск музыки",
	Help:    "Обеспечивает поиск исполнителей, альбомов, треков, позволяет перейти к загрузке",
	Factory: New,
}

func New(f connectivity.Interlayer, l logger.Logger, r registry.Registry) command.Command {
	return searchCommand{
		f: f,
		l: l.Fields(map[string]interface{}{"command": "search"}),
		r: r,
	}
}

func (s searchCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	const token = "b6c308fd-6a7f-441f-b120-a8d6e24126d9"
	cli, auth := s.f.Discovery.New(token)

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	req := music.SearchMusicParams{
		Limit:   utils.ToPointer(maxResult),
		Q:       arguments.String(),
		Context: ctx,
	}

	resp, err := cli.Music.SearchMusic(&req, auth)
	if err != nil {
		s.l.Logf(logger.ErrorLevel, "Search music failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	if len(resp.Payload.Results) == 0 {
		return messaging.NewSingleMessage(command.NothingFound, replyID)
	}

	messages := make([]messaging.ChatMessage, 0, len(resp.Payload.Results))
	for _, item := range resp.Payload.Results {
		msg := s.formatResult(item, replyID)
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	s.l.Logf(logger.InfoLevel, "Got %d results", len(messages))
	return messages
}
