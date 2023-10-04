package chatting

import "github.com/RacoonMediaServer/rms-music-bot/internal/messaging"

func (m *Manager) replyText(msg *messaging.Incoming, text string) {
	reply := &messaging.Outgoing{
		ChatID:  msg.ChatID,
		Message: messaging.New(text, msg.ID),
	}
	m.messenger.Outgoing() <- reply
}

func (m *Manager) replyRequestIsRequired(msg *messaging.Incoming) {
	reply := messaging.New("Для использования бота необходимо отправить заявку на предоставление доступа", msg.ID)
	reply.SetKeyboardStyle(messaging.MessageKeyboard)
	reply.AddButton("Подать заявку", "/request")
	m.messenger.Outgoing() <- &messaging.Outgoing{ChatID: msg.ChatID, Message: reply}
}

func (m *Manager) reply(msg *messaging.Incoming, replies []messaging.ChatMessage) {
	for i := range replies {
		m.messenger.Outgoing() <- &messaging.Outgoing{ChatID: msg.ChatID, Message: replies[i]}
	}
}
