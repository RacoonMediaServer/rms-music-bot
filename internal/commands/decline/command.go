package decline

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
)

type declineCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
}

var Command command.Type = command.Type{
	ID:      "decline",
	Title:   "Отклонение заявки",
	Help:    "Запретить доступ пользователя к боту",
	Factory: New,
	Attributes: command.Attributes{
		Internal:      true,
		AuthRequired:  true,
		AdminRequired: true,
	},
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return declineCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "decline"}),
	}
}

func (c declineCommand) Do(ctx command.Context) []messaging.ChatMessage {
	return messaging.NewSingleMessage("Заявка отклонена", ctx.ReplyID)
}
