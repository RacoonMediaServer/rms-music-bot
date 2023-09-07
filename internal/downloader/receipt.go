package downloader

import "github.com/anacrolix/torrent"

type receipt struct {
	t *torrent.Torrent
}

func (r receipt) Title() string {
	return r.t.Name()
}

func (r receipt) Wait() {
	<-r.t.GotInfo()
}
