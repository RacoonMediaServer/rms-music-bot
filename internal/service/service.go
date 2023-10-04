package service

import (
	"errors"
	"github.com/RacoonMediaServer/rms-music-bot/internal/access"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
	"sync"
	"time"
)

var errUnknownCommand = errors.New("unknown command")

const (
	resetAuthInterval = 12 * time.Hour
)

type Service struct {
	interlayer    connectivity.Interlayer
	accessService *access.Service

	mu    sync.RWMutex
	chats map[int]*chat
}

func New(interlayer connectivity.Interlayer) *Service {
	return &Service{
		interlayer:    interlayer,
		accessService: access.New(interlayer.Services, config.Config()),
		chats:         map[int]*chat{},
	}
}

func (s *Service) HandleMessage(messageID, userID int, userName, text string) []messaging.ChatMessage {
	s.mu.RLock()
	userChat, ok := s.chats[userID]
	s.mu.RUnlock()

	if !ok {
		userChat = s.addChat(userID)
	}

	probe, err := s.probeCommand(userChat, text)
	if err != nil {
		if errors.Is(err, errUnknownCommand) {
			logger.Warnf("handle '%s' from '%s' failed: %s", text, userName, errUnknownCommand)
			return messaging.NewSingleMessage("Нет такой команды", messageID)
		}
		logger.Errorf("handle command '%s' from '%s' failed: %s", text, userName, err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, messageID)
	}

	switch probe.accessVerdict {
	case accessGranted:
		ctx := command.Context{
			Arguments: probe.args,
			ReplyID:   messageID,
			Token:     probe.token,
			UserName:  userName,
			UserID:    userID,
		}
		return probe.cmd.Do(ctx)
	case accessNone:
		reply := messaging.New("Для использования бота необходимо отправить заявку на предоставление доступа", messageID)
		reply.SetKeyboardStyle(messaging.MessageKeyboard)
		reply.AddButton("Подать заявку", "/request")
		return []messaging.ChatMessage{reply}
	case accessDenied:
		return messaging.NewSingleMessage("Доступ к боту запрещен", messageID)
	case accessRequested:
		return messaging.NewSingleMessage("Заявка на доступ к боту отправлена. Ожидает одобрения администратором", messageID)
	default:
		return messaging.NewSingleMessage(command.SomethingWentWrong, messageID)
	}
}

func (s *Service) addChat(userId int) *chat {
	s.mu.Lock()
	defer s.mu.Unlock()

	userChat, ok := s.chats[userId]
	if ok {
		return userChat
	}

	userChat = &chat{
		lastCommand: "search",
		userId:      userId,
	}
	s.chats[userId] = userChat
	return userChat
}
