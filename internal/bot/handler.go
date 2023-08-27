package bot

import "github.com/RacoonMediaServer/rms-music-bot/internal/messaging"

type MessageHandler interface {
	HandleMessage(messageID, userID int, userName, text string) []messaging.ChatMessage
}
