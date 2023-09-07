package model

import "gorm.io/gorm"

type Artist struct {
	gorm.Model
	Name     string    `gorm:"index;unique;uniqueIndex"`
	Contents []Content `gorm:"foreignKey:ArtistID"`
}

type ContentType int

const (
	Discography ContentType = iota
	Album
)

type Content struct {
	gorm.Model
	Type     ContentType
	Title    string
	ArtistID uint
	Torrent  Torrent `gorm:"foreignKey:ContentID"`
}
