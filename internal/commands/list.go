package commands

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/add"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/artist"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/listen"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/play"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/search"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"go-micro.dev/v4/logger"
)

var commandMap map[string]command.Type

func init() {
	commandMap = map[string]command.Type{}
	commandMap[startCommandType.ID] = startCommandType
	commandMap[helpCommandType.ID] = helpCommandType

	commandMap[add.Command.ID] = add.Command
	commandMap[artist.Command.ID] = artist.Command
	commandMap[listen.Command.ID] = listen.Command
	commandMap[play.Command.ID] = play.Command
	commandMap[search.Command.ID] = search.Command
}

func NewCommand(commandID string, interlayer connectivity.Interlayer, l logger.Logger) (command.Command, command.Attributes, bool) {
	cmd, ok := commandMap[commandID]
	if !ok {
		return nil, command.Attributes{}, false
	}
	return cmd.Factory(interlayer, l), cmd.Attributes, true
}
