package add

import (
	"bytes"
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"go-micro.dev/v4/logger"
	"time"
)

type addCommand struct {
	f connectivity.Interlayer
	l logger.Logger
	r registry.Registry
}

const (
	maxResult     int64 = 10
	searchTimeout       = 1 * time.Minute
)

var Command command.Type = command.Type{
	ID:       "add",
	Title:    "Добавление музыки",
	Help:     "Добавляет музыку в библиотеку",
	Factory:  New,
	Internal: true,
}

func New(f connectivity.Interlayer, l logger.Logger, r registry.Registry) command.Command {
	return addCommand{
		f: f,
		l: l.Fields(map[string]interface{}{"command": "add"}),
		r: r,
	}
}

func (c addCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	if len(arguments) != 1 {
		return messaging.NewSingleMessage(command.ParseArgumentsFailed, replyID)
	}

	args, ok := registry.Get[*command.DownloadArguments](c.r, arguments[0])
	if !ok || !args.IsValid() {
		c.l.Logf(logger.WarnLevel, "Possible short link has expired")
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	q := args.Artist
	if args.Album != "" {
		q += " " + args.Album
	}

	const token = ""
	cli, auth := c.f.Discovery.New(token)

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	req := torrents.SearchTorrentsParams{
		Limit:   utils.ToPointer(maxResult),
		Q:       q,
		Strong:  utils.ToPointer(false),
		Type:    utils.ToPointer("music"),
		Context: ctx,
	}

	resp, err := cli.Torrents.SearchTorrents(&req, auth)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Search torrents failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	if len(resp.Payload.Results) == 0 {
		return messaging.NewSingleMessage(command.NothingFound, replyID)
	}

	selected := selectTorrent(resp.Payload.Results)
	downloadReq := torrents.DownloadTorrentParams{
		Link:    *selected.Link,
		Context: ctx,
	}

	buf := bytes.NewBuffer([]byte{})
	_, err = cli.Torrents.DownloadTorrent(&downloadReq, auth, buf)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Download torrent file failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	if err = c.f.Downloader.Download(buf.Bytes()); err != nil {
		c.l.Logf(logger.ErrorLevel, "Enqueue downloading failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	return messaging.NewSingleMessage("Добавлено", replyID)
}
