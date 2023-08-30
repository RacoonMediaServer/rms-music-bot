package db

import "github.com/RacoonMediaServer/rms-music-bot/internal/model"

func (d *Database) AddTorrent(t *model.Torrent) error {
	return d.conn.Create(t).Error
}

func (d *Database) LoadTorrents() (result []*model.Torrent, err error) {
	result = make([]*model.Torrent, 0)
	if err = d.conn.Find(&result).Error; err != nil {
		return nil, err
	}
	return
}
