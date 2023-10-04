package approve

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
)

type approveCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
}

var Command command.Type = command.Type{
	ID:      "approve",
	Title:   "Подтверждение заявки",
	Help:    "Разрешить доступ пользователя к боту",
	Factory: New,
	Attributes: command.Attributes{
		Internal:      true,
		AuthRequired:  true,
		AdminRequired: true,
	},
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return approveCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "approve"}),
	}
}

func (c approveCommand) Do(ctx command.Context) []messaging.ChatMessage {
	return messaging.NewSingleMessage("Заявка одобрена", ctx.ReplyID)
}
