package chatting

import (
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/commands"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
	"sync"
)

type Manager struct {
	messenger  Messenger
	interlayer connectivity.Interlayer

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu    sync.RWMutex
	chats map[int]*userChat
}

func (m *Manager) TriggerCommand(userID int, commandID string, args command.Arguments) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chat, ok := m.chats[userID]
	if !ok {
		return false
	}

	cmd, _, ok := commands.NewCommand(commandID, m.interlayer, logger.DefaultLogger)
	if !ok {
		return false
	}

	commandCtx := command.Context{
		Ctx:       m.ctx,
		Arguments: args,
		ReplyID:   0,
		Token:     chat.getToken(),
		Chat:      *chat.record,
		Chatting:  m,
	}
	m.reply(chat.record.ChatID, cmd.Do(commandCtx))
	return true
}

func (m *Manager) SendTo(userID int, message messaging.ChatMessage) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chat, ok := m.chats[userID]
	if !ok {
		return false
	}

	m.messenger.Outgoing() <- &messaging.Outgoing{ChatID: chat.record.ChatID, Message: message}
	return true
}

func (m *Manager) RequestAccess(userID int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chat, ok := m.chats[userID]
	if !ok {
		return false
	}

	chat.mu.Lock()
	defer chat.mu.Unlock()
	if chat.state != stateAccessGranted && chat.state != stateAccessDenied {
		chat.state = stateAccessRequested
		return true
	}
	return false
}

func NewManager(messenger Messenger, interlayer connectivity.Interlayer) *Manager {
	m := &Manager{
		messenger:  messenger,
		interlayer: interlayer,
		chats:      map[int]*userChat{},
	}
	if chats, err := interlayer.ChatStorage.LoadChats(); err == nil {
		for _, chatRecord := range chats {
			m.chats[chatRecord.UserID] = newUserChat(interlayer, chatRecord)
		}
	} else {
		logger.Errorf("Load stored chats failed: %s", err)
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())
	return m
}

func (m *Manager) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.loop()
	}()
}

func (m *Manager) loop() {
	for {
		select {
		case msg := <-m.messenger.Incoming():
			m.wg.Add(1)
			// TODO: job pool?
			go func(msg *messaging.Incoming) {
				defer m.wg.Done()
				m.processIncomingMessage(msg)
			}(msg)
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) processIncomingMessage(msg *messaging.Incoming) {
	chat := m.getOrCreateChat(msg)
	state, err := chat.requestState()
	if err != nil {
		logger.Errorf("Request chat state failed: %s", err)
		m.replyText(msg, command.SomethingWentWrong)
		return
	}

	if state == stateAccessDenied {
		m.replyText(msg, "Доступ к боту запрещен")
		return
	}
	if state == stateAccessRequested {
		m.replyText(msg, "Заявка на доступ к боту отправлена. Ожидает одобрения администратором")
		return
	}

	text := msg.Text
	if !command.IsCommand(msg.Text) {
		text = fmt.Sprintf("/%s %s", chat.loadPrevCommand(), text)
	}

	commandID, commandArgs := command.Parse(text)
	cmd, attrs, ok := commands.NewCommand(commandID, m.interlayer, logger.DefaultLogger)
	if !ok {
		m.replyText(msg, "Нет такой команды. Список команд описан в /help")
		return
	}

	if attrs.AuthRequired && state != stateAccessGranted {
		m.replyRequestIsRequired(msg)
		return
	}
	if attrs.CanRepeat {
		chat.savePrevCommand(commandID)
	}

	commandCtx := command.Context{
		Ctx:       m.ctx,
		Arguments: commandArgs,
		ReplyID:   msg.ID,
		Token:     chat.getToken(),
		Chat:      *chat.record,
		Chatting:  m,
	}
	m.reply(msg.ChatID, cmd.Do(commandCtx))
}

func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
}
