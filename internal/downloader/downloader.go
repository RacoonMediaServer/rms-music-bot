package downloader

import (
	"bytes"
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"github.com/anacrolix/fuse"
	fusefs "github.com/anacrolix/fuse/fs"
	"github.com/anacrolix/torrent"
	torrentfs "github.com/anacrolix/torrent/fs"
	"github.com/anacrolix/torrent/metainfo"
	"go-micro.dev/v4/logger"
	"os"
	"path"
	"sync"
)

type Downloader struct {
	layout config.Layout
	db     Database
	cli    *torrent.Client
	l      logger.Logger
	fs     *torrentfs.TorrentFS
	wg     sync.WaitGroup

	// TODO: сделать отдельный глобальный режим для обслуживания, лочка - плохо
	mu sync.RWMutex
}

func init() {
	fuse.Debug = func(msg interface{}) {
		logger.Infof("fuse: %s", msg)
	}
}

func New(layout config.Layout, db Database) *Downloader {
	return &Downloader{
		layout: layout,
		db:     db,
		l:      logger.Fields(map[string]interface{}{"from": "downloader"}),
	}
}

func (d *Downloader) Start() error {
	torrents, err := d.db.LoadTorrents()
	if err != nil {
		return fmt.Errorf("load torrents failed: %w", err)
	}
	conn, err := fuse.Mount(d.layout.Directory, fuse.AllowOther())
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

	d.l.Log(logger.InfoLevel, "Loading stored torrents...")
	for _, t := range torrents {
		if _, err = d.registerTorrent(t.Bytes); err != nil {
			d.l.Logf(logger.WarnLevel, "Load '%s' failed: %s", t.Title, err)
		}
	}
	d.l.Log(logger.InfoLevel, "Ready")

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

func (d *Downloader) registerTorrent(content []byte) (*torrent.Torrent, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var spec *torrent.TorrentSpec
	isMagnet := isMagnetLink(content)
	if !isMagnet {
		mi, err := metainfo.Load(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}
		spec = torrent.TorrentSpecFromMetaInfo(mi)
	} else {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(string(content))
		if err != nil {
			return nil, err
		}
	}

	opts := torrent.AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		ChunkSize: spec.ChunkSize,
	}

	t, _ := d.cli.AddTorrentOpt(opts)
	if err := t.MergeSpec(spec); err != nil {
		t.Drop()
		return nil, err
	}

	return t, nil
}

func (d *Downloader) Add(content []byte) (model.Download, error) {
	t, err := d.registerTorrent(content)
	if err != nil {
		return nil, err
	}

	return &receipt{t: t}, nil
}

func (d *Downloader) GetFile(filepath string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return os.ReadFile(path.Join(d.layout.Directory, filepath))
}

func (d *Downloader) Wipe() {
	d.mu.Lock()

	d.l.Logf(logger.InfoLevel, "Wiping...")

	curTorrents := d.cli.Torrents()
	for _, t := range curTorrents {
		t.Drop()
	}
	files, err := os.ReadDir(d.layout.Downloads)
	if err != nil {
		d.l.Logf(logger.FatalLevel, "Get downloaded files failed: %s", err)
	}
	for _, f := range files {
		_ = os.RemoveAll(path.Join(d.layout.Downloads, f.Name()))
	}
	d.mu.Unlock()

	torrents, err := d.db.LoadTorrents()
	if err != nil {
		d.l.Logf(logger.FatalLevel, "Load torrents failed: %s", err)
	}
	for _, t := range torrents {
		if _, err = d.registerTorrent(t.Bytes); err != nil {
			d.l.Logf(logger.WarnLevel, "Load '%s' failed: %s", t.Title)
		}
	}

	d.l.Logf(logger.InfoLevel, "Wipe successfully done")
}

func (d *Downloader) Stop() {
	d.fs.Destroy()
	if err := fuse.Unmount(d.layout.Directory); err != nil {
		d.l.Logf(logger.ErrorLevel, "Umount failed: %s", err)
	}
	d.wg.Wait()
}
