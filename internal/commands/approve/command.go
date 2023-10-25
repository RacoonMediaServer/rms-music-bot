package approve

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	rms_users "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-users"
	"github.com/delucks/go-subsonic"
	"go-micro.dev/v4/logger"
	"net/http"
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

	if err = c.registerServiceUser(req); err != nil {
		c.l.Logf(logger.ErrorLevel, "Register user to music service failed: %s", err)
		return messaging.NewSingleMessage("Возникла ошибка регистрации пользователя на стримере", ctx.ReplyID)
	}

	ctx.Chatting.SendTo(req.UserID, messaging.New("Заявка на доступ к боту одобрена!", 0))
	ctx.Chatting.TriggerCommand(req.UserID, "help", command.Arguments{})

	return messaging.NewSingleMessage("Заявка одобрена", ctx.ReplyID)
}

func (c approveCommand) registerServiceUser(req *command.Request) error {
	conf := config.Config().Service
	cli := subsonic.Client{
		Client:       &http.Client{},
		BaseUrl:      conf.Server,
		User:         conf.Username,
		ClientName:   "music-bot",
		PasswordAuth: true,
	}
	if err := cli.Authenticate(conf.Password); err != nil {
		return err
	}

	return cli.CreateUser(req.UserName, req.Password, "", map[string]string{})
}
