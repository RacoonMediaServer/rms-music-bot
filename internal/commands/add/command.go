package add

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"go-micro.dev/v4/logger"
	"time"
)

type addCommand struct {
	f connectivity.Factory
	l logger.Logger
	r registry.Registry
}

const (
	maxResult     int64 = 4
	searchTimeout       = 30 * time.Second
)

var Command command.Type = command.Type{
	ID:       "add",
	Title:    "Добавление музыки",
	Help:     "Добавляет музыку в библиотеку",
	Factory:  New,
	Internal: true,
}

func New(f connectivity.Factory, l logger.Logger, r registry.Registry) command.Command {
	return addCommand{
		f: f,
		l: l.Fields(map[string]interface{}{"command": "add"}),
		r: r,
	}
}

func (a addCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
}
