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
}

type helpCommand struct {
}

func (h helpCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	titles := make([]string, 0, len(commandMap))
	for k, _ := range commandMap {
		titles = append(titles, k)
	}
	sort.Slice(titles, func(i, j int) bool {
		return titles[i] < titles[j]
	})
	result := `Данные бот предназначен для просулшивания музыки без цензуры, <strike>регистрации и СМС</strike>. Музыку можно слушать через Telegram, с телефона или <a href="https://music.racoondev.top/">веб-сайта</a>. Для прослушивания музыки с телефона - необходимо установить приложение <a href="https://play.google.com/store/apps/details?id=com.ghenry22.substream2">substreamer</a>. Данные для выхода:\n`
	result += `Server: https://music.racoondev.top/\n`
	result += `Username: demo\n`
	result += `Password: demo\n\n`

	for _, t := range titles {
		cmd := commandMap[t]
		if !cmd.Internal {
			result += fmt.Sprintf("/%s %s - %s\n", cmd.ID, cmd.Title, cmd.Help)
		}
	}

	result += "Весь функционал на данный момент является <b>демо-версией</b>"
	return messaging.NewSingleMessage(result, replyID)
}

func newHelpCommand(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return &helpCommand{}
}
