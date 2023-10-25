package chatting

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"go-micro.dev/v4/logger"
)

func (m *Manager) getOrCreateChat(msg *messaging.Incoming) *userChat {
	m.mu.RLock()
	chat, ok := m.chats[msg.UserID]
	m.mu.RUnlock()

	if !ok {
		chat = m.createChat(msg)
		if err := m.interlayer.ChatStorage.SaveChat(&model.Chat{ChatID: msg.ChatID, UserID: msg.UserID}); err != nil {
			logger.Errorf("Save chat record failed: %s", err)
		}
	}
	return chat
}

func (m *Manager) createChat(msg *messaging.Incoming) *userChat {
	m.mu.Lock()
	defer m.mu.Unlock()

	chat, ok := m.chats[msg.UserID]
	if !ok {
		chat = newUserChat(m.interlayer, msg.ChatID, msg.UserID)
		m.chats[msg.UserID] = chat
	}
	return chat
}
