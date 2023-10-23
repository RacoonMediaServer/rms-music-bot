package approve

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
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
	if len(ctx.Arguments) != 1 {
		return messaging.NewSingleMessage(command.ParseArgumentsFailed, ctx.ReplyID)
	}
	req, ok := registry.Get[*command.Request](c.interlayer.Registry, ctx.Arguments[0])
	if !ok {
		return messaging.NewSingleMessage("Заявка не найдена, либо просрочена", ctx.ReplyID)
	}

	userID := int32(req.UserID)
	user := &rms_users.User{
		Name:           req.UserName,
		TelegramUserID: &userID,
		Perms: []rms_users.Permissions{
			rms_users.Permissions_Search,
			rms_users.Permissions_ListeningMusic,
		},
	}

	_, err := c.interlayer.Services.NewUsers().RegisterUser(ctx.Ctx, user)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Register user failed: %s", err)
		return messaging.NewSingleMessage("Не удалось зарегистрировать нового пользователя", ctx.ReplyID)
	}

	ctx.Chatting.SendTo(req.UserID, messaging.New("Заявка на доступ к боту одобрена! Для справки по командам можно использовать /help.", 0))

	//TODO: register service user
	return messaging.NewSingleMessage("Заявка одобрена", ctx.ReplyID)
}
