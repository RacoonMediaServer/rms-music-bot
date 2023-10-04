package request

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
)

type requestCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
}

var Command command.Type = command.Type{
	ID:      "request",
	Title:   "Отправка заявки",
	Help:    "Позволяет отправить заявку на предоставление доступа к боту",
	Factory: New,
	Attributes: command.Attributes{
		Internal: true,
	},
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return requestCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "request"}),
	}
}

func (c requestCommand) Do(ctx command.Context) []messaging.ChatMessage {
	return messaging.NewSingleMessage("Заявка отправлена", ctx.ReplyID)
}
