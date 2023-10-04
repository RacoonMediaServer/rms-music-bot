package service

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands"
	"go-micro.dev/v4/logger"
	"sync"
	"time"
)

type accessState int

const (
	accessGranted accessState = iota
	accessNone
	accessDenied
	accessRequested
)

type chat struct {
	mu              sync.Mutex
	userId          int
	access          accessState
	token           string
	checkAccessTime time.Time
	lastCommand     string
}

type probeResult struct {
	cmd           command.Command
	args          command.Arguments
	token         string
	accessVerdict accessState
}

func (s *Service) probeCommand(userChat *chat, text string) (probeResult, error) {
	userChat.mu.Lock()
	defer userChat.mu.Unlock()
	result := probeResult{}

	if !command.IsCommand(text) {
		text = fmt.Sprintf("/%s %s", userChat.lastCommand, text)
	}

	if userChat.access != accessGranted || time.Now().Sub(userChat.checkAccessTime) > resetAuthInterval {
		granted, token, err := s.accessService.CheckAccess(userChat.userId)
		if err != nil {
			return result, fmt.Errorf("check access failed: %s", err)
		}
		if granted {
			userChat.access = accessGranted
			userChat.token = token
			userChat.checkAccessTime = time.Now()
		} else if userChat.access == accessGranted {
			userChat.access = accessDenied
		}
	}

	commandID, args := command.Parse(text)
	cmd, attrs, ok := commands.NewCommand(commandID, s.interlayer, logger.DefaultLogger)
	if !ok {
		logger.Warnf("command '%s' not found", commandID)
		return result, errUnknownCommand
	}
	if attrs.CanRepeat {
		userChat.lastCommand = commandID
	}

	result.cmd = cmd
	result.args = args
	result.token = userChat.token
	result.accessVerdict = userChat.access
	if userChat.access != accessGranted && !attrs.AuthRequired {
		result.accessVerdict = accessGranted
	}
	return result, nil
}
