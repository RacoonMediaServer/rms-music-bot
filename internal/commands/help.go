package commands

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
	"sort"
)

var helpCommandType = command.Type{
	ID:      "help",
	Title:   "Справка",
	Help:    "Пояснить за функции бота",
	Factory: newHelpCommand,
	Attributes: command.Attributes{
		AuthRequired: true,
	},
}

var startCommandType = command.Type{
	ID:      "start",
	Title:   "Начало",
	Help:    "Начать работу с ботом",
	Factory: newHelpCommand,
	Attributes: command.Attributes{
		Internal:     true,
		AuthRequired: true,
	},
}

type helpCommand struct {
}

func (h helpCommand) Do(ctx command.Context) []messaging.ChatMessage {
	titles := make([]string, 0, len(commandMap))
	for k, _ := range commandMap {
		titles = append(titles, k)
	}
	sort.Slice(titles, func(i, j int) bool {
		return titles[i] < titles[j]
	})
	result := `Данный бот предназначен для прослушивания музыки без цензуры, <strike>регистрации и СМС</strike>. Дискография исполнителей добавляется с помощью нижеперечисленных команд, музыку же можно воспроизвести через Telegram, с телефона или <a href="https://music.racoondev.top/">веб-сайта</a>. Для прослушивания музыки с телефона - необходимо установить приложение <a href="https://play.google.com/store/apps/details?id=com.ghenry22.substream2">substreamer</a>.`
	result += "\n\nServer: https://music.racoondev.top/\n"
	result += "Username: <b>demo</b>\n"
	result += "Password: <b>demo</b>\n\n"

	for _, t := range titles {
		cmd := commandMap[t]
		if !cmd.Attributes.Internal {
			result += fmt.Sprintf("/%s %s - %s\n", cmd.ID, cmd.Title, cmd.Help)
		}
	}

	result += "\nВесь функционал на данный момент является <b>демо-версией</b>."
	return messaging.NewSingleMessage(result, ctx.ReplyID)
}

func newHelpCommand(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return &helpCommand{}
}
