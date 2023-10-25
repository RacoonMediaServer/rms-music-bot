package commands

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/formatter"
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

	helpText := ""
	for _, t := range titles {
		cmd := commandMap[t]
		if !cmd.Attributes.Internal {
			helpText += fmt.Sprintf("/%s %s - %s\n", cmd.ID, cmd.Title, cmd.Help)
		}
	}

	helpCtx := formatter.HelpContext{
		Link:     config.Config().Service.Server,
		UserName: ctx.Chat.UserName,
		Password: ctx.Chat.Password,
		Text:     helpText,
	}

	return []messaging.ChatMessage{formatter.FormatHelp(helpCtx, ctx.ReplyID)}
}

func newHelpCommand(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return &helpCommand{}
}
