package service

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
)

type Service struct {
	interlayer connectivity.Interlayer
}

func New(interlayer connectivity.Interlayer) *Service {
	return &Service{interlayer: interlayer}
}

func (s Service) HandleMessage(messageID, userID int, userName, text string) []messaging.ChatMessage {
	if !command.IsCommand(text) {
		text = "/search " + text
	}
	commandID, args := command.Parse(text)
	cmd, err := commands.NewCommand(commandID, s.interlayer, logger.DefaultLogger)
	if err != nil {
		logger.Warnf("cannot execute command '%s': %s", commandID, err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, messageID)
	}
	return cmd.Do(args, messageID)
}
