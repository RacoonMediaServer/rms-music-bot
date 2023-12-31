package request

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
	"time"
)

const requestTTL = 24 * time.Hour

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
	if !ctx.Chatting.RequestAccess(ctx.Chat.UserID) {
		return messaging.NewSingleMessage("Не удалось отправить заявку", ctx.ReplyID)
	}

	req := command.Request{
		UserID:   ctx.Chat.UserID,
		UserName: ctx.Chat.UserName,
		Password: ctx.Chat.Password,
	}
	reqId := c.interlayer.Registry.Add(&req, requestTTL)

	text := fmt.Sprintf("Заявка на доступ к боту от нового пользователя:\n\nИмя: <b>%s</b>\nИдентификатор: <b>%d</b>\nПароль: <b>%s</b>\n",
		ctx.Chat.UserName, ctx.Chat.UserID, ctx.Chat.Password)
	msg := messaging.New(text, 0)
	msg.SetKeyboardStyle(messaging.MessageKeyboard)
	msg.AddButton("Принять", "/approve "+reqId)
	msg.AddButton("Отклонить", "/decline "+reqId)

	adminUsers, err := c.interlayer.AccessService.GetAdminUserId()
	if err != nil {
		c.l.Logf(logger.WarnLevel, "Cannot send request to admin user: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, ctx.ReplyID)
	}

	requestSent := false
	for _, adminUserId := range adminUsers {
		requestSent = ctx.Chatting.SendTo(adminUserId, msg)
		if requestSent {
			break
		}
	}

	if !requestSent {
		return messaging.NewSingleMessage("Возникли проблемы с запросом заявки. Стоит попробовать позже", ctx.ReplyID)
	}

	return messaging.NewSingleMessage("Заявка отправлена", ctx.ReplyID)
}
