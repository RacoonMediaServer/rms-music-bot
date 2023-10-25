package model

import (
	"fmt"
	"github.com/sethvargo/go-password/password"
)

type Chat struct {
	UserID   int    `gorm:"primaryKey"`
	ChatID   int64  `gorm:"unique"`
	UserName string `gorm:"unique"`
	Password string
}

func (chat *Chat) GeneratePassword() error {
	pw, err := password.Generate(8, 3, 0, false, true)
	if err != nil {
		return err
	}

	chat.Password = pw
	return nil
}

func NewChat(userID int, chatID int64, userName string) *Chat {
	chat := Chat{
		UserID:   userID,
		ChatID:   chatID,
		UserName: userName,
	}
	if chat.UserName == "" {
		chat.UserName = fmt.Sprintf("%d", userID)
	}
	if err := chat.GeneratePassword(); err != nil {
		panic(fmt.Errorf("generate password failed: %w", err))
	}
	return &chat
}
