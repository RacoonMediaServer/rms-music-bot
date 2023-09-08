package db

import (
	"errors"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"gorm.io/gorm"
)

func (d *Database) AddContent(artistName string, content model.Content) error {
	var artist model.Artist
	if err := d.conn.Model(&artist).Preload("Contents").Preload("Contents.Torrent").First(&artist, "name = ?", artistName).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			artist.Name = artistName
			artist.Contents = append(artist.Contents, content)
			return d.conn.Create(&artist).Error
		}
		return err
	}
	artist.Contents = append(artist.Contents, content)
	return d.conn.Save(&artist).Error
}

func (d *Database) GetContent(artistName string) ([]model.Content, error) {
	var artist model.Artist
	err := d.conn.Model(&artist).Preload("Contents").Preload("Contents.Torrent").First(&artist, "name = ?", artistName).Error
	return artist.Contents, err

}
