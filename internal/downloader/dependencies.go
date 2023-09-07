package downloader

import "github.com/RacoonMediaServer/rms-music-bot/internal/model"

type Database interface {
	LoadTorrents() ([]*model.Torrent, error)
}
