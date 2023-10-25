package model

type Chat struct {
	UserID int   `gorm:"primaryKey"`
	ChatID int64 `gorm:"unique"`
}
