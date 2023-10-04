package health

import (
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/delucks/go-subsonic"
	"go-micro.dev/v4/logger"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	checkInterval   = 15 * time.Minute
	readPieceSize   = 16 * 1024
	downloadTimeout = 60 * time.Second
)

var errDownloadIsHung = errors.New("download is hung")

type Checker struct {
	healthCheckConfig config.HealthCheck
	serviceConfig     config.Service
	wg                sync.WaitGroup
	l                 logger.Logger
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewChecker(conf config.Configuration) *Checker {
	chk := Checker{
		healthCheckConfig: conf.HealthCheck,
		serviceConfig:     conf.Service,
		l:                 logger.Fields(map[string]interface{}{"from": "healthcheck"}),
	}
	chk.ctx, chk.cancel = context.WithCancel(context.Background())
	if conf.HealthCheck.Enabled {
		chk.startAsyncCheck()
	}
	return &chk
}

func (c *Checker) startAsyncCheck() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.process()
	}()
}

func (c *Checker) process() {
	t := time.NewTicker(checkInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			c.startSyncCheck()
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Checker) startSyncCheck() {
	c.l.Logf(logger.DebugLevel, "Checking music service....")
	if err := c.check(); err != nil {
		if errors.Is(err, errDownloadIsHung) {
			c.l.Logf(logger.ErrorLevel, "Music service is not healthy: %s", err)
			c.notifyFifo()
		} else {
			c.l.Logf(logger.WarnLevel, "Cannot check music service: %s", err)
		}
		return
	}
	c.l.Logf(logger.DebugLevel, "Music service: OK")
}

func (c *Checker) check() error {
	cli := subsonic.Client{
		Client:       &http.Client{Timeout: downloadTimeout},
		BaseUrl:      c.serviceConfig.Server,
		User:         c.serviceConfig.Username,
		ClientName:   "music-bot",
		PasswordAuth: true,
	}
	if err := cli.Authenticate(c.serviceConfig.Password); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	albums, err := cli.GetAlbumList("random", map[string]string{"size": "1"})
	if err != nil {
		return fmt.Errorf("get random album failed: %w", err)
	}
	if len(albums) == 0 {
		return errors.New("cannot get any random album")
	}
	album, err := cli.GetAlbum(albums[0].ID)
	if err != nil {
		return fmt.Errorf("get album failed: %w", err)
	}
	if len(album.Song) == 0 {
		return errors.New("no any songs in random album")
	}

	rd, err := getStream(&cli, album.Song[0].ID, map[string]string{})
	if err != nil {
		return fmt.Errorf("%w: download failed: %s", errDownloadIsHung, err)
	}
	defer rd.Close()

	buf := make([]byte, readPieceSize)
	_, err = io.ReadFull(rd, buf)
	if err != nil {
		return fmt.Errorf("%w: read failed: %s", errDownloadIsHung, err)
	}
	return nil
}

func (c *Checker) notifyFifo() {
	f, err := os.OpenFile(c.healthCheckConfig.Fifo, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Open Fifo failed: %s", err)
		return
	}
	_, _ = f.Write([]byte{'0'})
	_ = f.Close()
}

func (c *Checker) Stop() {
	c.cancel()
	c.wg.Wait()
}
