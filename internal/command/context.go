package command

import (
	"context"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
)

type ChatSystem interface {
	SendTo(userID int, message messaging.ChatMessage) bool
	TriggerCommand(userID int, command string, args Arguments) bool
	RequestAccess(userID int) bool
}

type Context struct {
	Ctx       context.Context
	Arguments Arguments
	ReplyID   int
	Token     string
	Chat      model.Chat
	Chatting  ChatSystem
}
