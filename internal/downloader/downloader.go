package downloader

import (
	"fmt"
	tConfig "github.com/RacoonMediaServer/distribyted/config"
	"github.com/RacoonMediaServer/distribyted/fuse"
	"github.com/RacoonMediaServer/distribyted/torrent"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/anacrolix/missinggo/v2/filecache"
	aTorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"go-micro.dev/v4/logger"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

const mainRoute = "music"
const cacheTTL = 1 * time.Hour
const readTimeout = 60
const addTimeout = 60

type Downloader struct {
	layout config.Layout
	db     Database
	l      logger.Logger
	wg     sync.WaitGroup

	fuse      *fuse.Handler
	fileStore *torrent.FileItemStore
	cli       *aTorrent.Client
	service   *torrent.Service

	// TODO: сделать отдельный глобальный режим для обслуживания, лочка - плохо
	mu sync.RWMutex
}

func New(layout config.Layout, db Database) *Downloader {
	return &Downloader{
		layout: layout,
		db:     db,
		l:      logger.Fields(map[string]interface{}{"from": "downloader"}),
	}
}

func (d *Downloader) Start() error {
	if err := os.MkdirAll(d.layout.Directory, 0744); err != nil {
		return fmt.Errorf("create content directory failed: %w", err)
	}
	if err := os.MkdirAll(d.layout.Downloads, 0744); err != nil {
		return fmt.Errorf("create downloads directory failed: %w", err)
	}
	pieceCompletionDir := filepath.Join(d.layout.Downloads, "piece-completion")
	if err := os.MkdirAll(pieceCompletionDir, 0744); err != nil {
		return fmt.Errorf("create piece completion directory failed: %w", err)
	}
	fcache, err := filecache.NewCache(filepath.Join(d.layout.Downloads, "cache"))
	if err != nil {
		return fmt.Errorf("create cache failed: %w", err)
	}
	fcache.SetCapacity(int64(d.layout.Limit) * 1024 * 1024 * 1024)

	torrentStorage := storage.NewResourcePieces(fcache.AsResourceProvider())

	fileStore, err := torrent.NewFileItemStore(filepath.Join(d.layout.Downloads, "items"), cacheTTL)
	if err != nil {
		return fmt.Errorf("create file store failed: %w", err)
	}

	id, err := torrent.GetOrCreatePeerID(filepath.Join(d.layout.Downloads, "ID"))
	if err != nil {
		return fmt.Errorf("create ID failed: %w", err)
	}

	conf := tConfig.TorrentGlobal{
		ReadTimeout:     readTimeout,
		AddTimeout:      addTimeout,
		GlobalCacheSize: -1,
		MetadataFolder:  d.layout.Downloads,
		DisableIPv6:     false,
	}

	cli, err := torrent.NewClient(torrentStorage, fileStore, &conf, id)
	if err != nil {
		return fmt.Errorf("start torrent client failed: %w", err)
	}

	//pieceCompletionStorage, err := storage.NewBoltPieceCompletion(pieceCompletionDir)
	//if err != nil {
	//	return fmt.Errorf("create piece completion storage failed: %w", err)
	//}

	stats := torrent.NewStats()

	loaders := []torrent.DatabaseLoader{&loader{db: d.db}}
	service := torrent.NewService(loaders, stats, cli, conf.AddTimeout, conf.ReadTimeout)

	fss, err := service.Load()
	if err != nil {
		return fmt.Errorf("load torrents failed: %w", err)
	}

	mh := fuse.NewHandler(true, d.layout.Directory)
	if err = mh.Mount(fss); err != nil {
		return fmt.Errorf("mount fuse directory: %w", err)
	}

	d.fuse = mh
	d.fileStore = fileStore
	d.cli = cli
	d.service = service

	return nil
}

func (d *Downloader) Add(content []byte) (string, error) {
	title, err := d.service.Add(mainRoute, content)
	if err != nil {
		return "", err
	}
	return title, nil
}

func (d *Downloader) GetFile(filepath string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return os.ReadFile(path.Join(d.layout.Directory, mainRoute, filepath))
}

func (d *Downloader) Wipe() {
	d.l.Logf(logger.InfoLevel, "Wiping...")

	// TODO

	d.l.Logf(logger.InfoLevel, "Wipe successfully done")
}

func (d *Downloader) Stop() {
	_ = d.fileStore.Close()
	d.cli.Close()
	d.fuse.Unmount()
}
