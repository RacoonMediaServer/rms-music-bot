package commands

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
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
	result := ""
	for _, t := range titles {
		cmd := commandMap[t]
		if !cmd.Internal {
			result += fmt.Sprintf("/%s %s - %s\n", cmd.ID, cmd.Title, cmd.Help)
		}
	}
	return messaging.NewSingleMessage(result, replyID)
}

func newHelpCommand(factory connectivity.Interlayer, l logger.Logger, r registry.Registry) command.Command {
	return &helpCommand{}
}
