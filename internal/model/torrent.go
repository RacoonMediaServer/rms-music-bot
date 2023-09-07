package model

import "gorm.io/gorm"

type Torrent struct {
	gorm.Model
	Title     string
	ContentID uint
	Bytes     []byte
}
