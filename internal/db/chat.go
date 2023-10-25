package db

import (
	"errors"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"gorm.io/gorm"
)

func (d *Database) LoadChats() (result []*model.Chat, err error) {
	result = make([]*model.Chat, 0)
	if err = d.conn.Find(&result).Error; err != nil {
		return nil, err
	}
	return
}

func (d *Database) SaveChat(chat *model.Chat) error {
	foundChat := model.Chat{}
	if err := d.conn.First(&foundChat, "user_id = ?", chat.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return d.conn.Save(chat).Error
		}
		return err
	}
	return d.conn.Create(chat).Error
}
