package command

import (
	"context"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
)

type ChatSystem interface {
	SendTo(userID int, message messaging.ChatMessage) bool
	RequestAccess(userID int) bool
}

type Context struct {
	Ctx       context.Context
	Arguments Arguments
	ReplyID   int
	Token     string
	UserName  string
	UserID    int
	Chatting  ChatSystem
}
