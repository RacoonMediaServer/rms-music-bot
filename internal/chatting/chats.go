package chatting

import "github.com/RacoonMediaServer/rms-music-bot/internal/messaging"

func (m *Manager) getOrCreateChat(msg *messaging.Incoming) *userChat {
	m.mu.RLock()
	chat, ok := m.chats[msg.UserID]
	m.mu.RUnlock()

	if !ok {
		chat = m.createChat(msg)
	}
	return chat
}

func (m *Manager) createChat(msg *messaging.Incoming) *userChat {
	m.mu.Lock()
	defer m.mu.Unlock()

	chat, ok := m.chats[msg.UserID]
	if !ok {
		chat = &userChat{
			accessService: m.interlayer.AccessService,
			chatID:        msg.ChatID,
			userID:        msg.UserID,
			prevCommand:   "search",
		}
		m.chats[msg.UserID] = chat
	}
	return chat
}
