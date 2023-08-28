package commands

import (
	"errors"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/add"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands/search"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"go-micro.dev/v4/logger"
)

var commandMap map[string]command.Type

var ErrCommandNotFound = errors.New("unknown command")

func init() {
	commandMap = map[string]command.Type{}
	commandMap[helpCommandType.ID] = helpCommandType
	commandMap[search.Command.ID] = search.Command
	commandMap[add.Command.ID] = add.Command
}

func NewCommand(commandID string, f connectivity.Interlayer, l logger.Logger, r registry.Registry) (command.Command, error) {
	cmd, ok := commandMap[commandID]
	if !ok {
		return nil, ErrCommandNotFound
	}
	return cmd.Factory(f, l, r), nil
}
