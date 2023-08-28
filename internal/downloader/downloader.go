package downloader

import (
	"bytes"
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/anacrolix/fuse"
	fusefs "github.com/anacrolix/fuse/fs"
	"github.com/anacrolix/torrent"
	torrentfs "github.com/anacrolix/torrent/fs"
	"github.com/anacrolix/torrent/metainfo"
	"go-micro.dev/v4/logger"
	"sync"
)

type Downloader struct {
	layout config.Layout
	cli    *torrent.Client
	l      logger.Logger
	fs     *torrentfs.TorrentFS
	wg     sync.WaitGroup
}

func New(layout config.Layout) *Downloader {
	return &Downloader{
		layout: layout,
		l:      logger.Fields(map[string]interface{}{"from": "downloader"}),
	}
}

func (d *Downloader) Start() error {
	conn, err := fuse.Mount(d.layout.Directory)
	if err != nil {
		return fmt.Errorf("mount fuse dir failed: %w", err)
	}

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = d.layout.Downloads
	cfg.NoUpload = true // Ensure that downloads are responsive.
	d.cli, err = torrent.NewClient(cfg)
	if err != nil {
		_ = fuse.Unmount(d.layout.Directory)
		return fmt.Errorf("create torrent client failed: %w", err)
	}

	d.fs = torrentfs.New(d.cli)

	d.wg.Add(1)
	go d.serveFS(conn)

	return nil
}

func (d *Downloader) serveFS(conn *fuse.Conn) {
	defer d.wg.Done()

	if err := fusefs.Serve(conn, d.fs); err != nil {
		d.l.Logf(logger.ErrorLevel, "Serve filesystem failed: %s", err)
		return
	}
	<-conn.Ready
	if err := conn.MountError; err != nil {
		d.l.Logf(logger.ErrorLevel, "Mount error: %s", err)
	}
}

func (d *Downloader) Download(content []byte) error {
	var spec *torrent.TorrentSpec
	isMagnet := isMagnetLink(content)
	if !isMagnet {
		mi, err := metainfo.Load(bytes.NewReader(content))
		if err != nil {
			return err
		}
		spec = torrent.TorrentSpecFromMetaInfo(mi)
	} else {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(string(content))
		if err != nil {
			return err
		}
	}

	opts := torrent.AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		ChunkSize: spec.ChunkSize,
	}

	t, _ := d.cli.AddTorrentOpt(opts)
	if err := t.MergeSpec(spec); err != nil {
		t.Drop()
		return nil
	}

	return nil
}

func (d *Downloader) Stop() {
	d.fs.Destroy()
	if err := fuse.Unmount(d.layout.Directory); err != nil {
		d.l.Logf(logger.ErrorLevel, "Umount failed: %s", err)
	}
	d.wg.Wait()
}
