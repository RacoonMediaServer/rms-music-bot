package add

import (
	"bytes"
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-music-bot/internal/selector"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"github.com/go-openapi/runtime"
	"go-micro.dev/v4/logger"
	"path"
	"strings"
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

	allAlbums := args.Album == "" && args.Track == ""

	const token = "b6c308fd-6a7f-441f-b120-a8d6e24126d9"
	cli, auth := c.f.Discovery.New(token)
	var variants []*models.SearchTorrentsResult
	var err error

	for len(variants) == 0 {
		variants, err = c.searchTorrents(cli, auth, q, allAlbums)
		if err != nil {
			c.l.Logf(logger.ErrorLevel, "Search torrents failed: %s", err)
			return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
		}
		if len(variants) == 0 {
			if !allAlbums {
				allAlbums = true
				q = args.Artist
			} else {
				return messaging.NewSingleMessage(command.NothingFound, replyID)
			}
		}
	}

	sel := selector.MusicSelector{
		Query:               q,
		Discography:         allAlbums,
		MinSeedersThreshold: 10,
		MaxSizeMB:           50 * 1024,
		Format:              "mp3",
	}
	chosen := sel.Select(variants)

	files, err := c.download(cli, auth, *chosen.Link)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Enqueue downloading failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	if args.Track != "" {
		return c.play(args.Track, files, replyID)
	}

	return messaging.NewSingleMessage("Добавлено", replyID)
}

func (c addCommand) searchTorrents(cli *client.Client, auth runtime.ClientAuthInfoWriter, q string, allAlbums bool) ([]*models.SearchTorrentsResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	req := torrents.SearchTorrentsParams{
		Limit:       utils.ToPointer(maxResult),
		Q:           q,
		Strong:      utils.ToPointer(false),
		Type:        utils.ToPointer("music"),
		Discography: utils.ToPointer(allAlbums),
		Context:     ctx,
	}

	resp, err := cli.Torrents.SearchTorrents(&req, auth)
	if err != nil {
		return nil, err
	}
	return resp.Payload.Results, nil
}

func (c addCommand) download(cli *client.Client, auth runtime.ClientAuthInfoWriter, link string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	downloadReq := torrents.DownloadTorrentParams{
		Link:    link,
		Context: ctx,
	}

	buf := bytes.NewBuffer([]byte{})
	_, err := cli.Torrents.DownloadTorrent(&downloadReq, auth, buf)
	if err != nil {
		return nil, err
	}

	return c.f.Downloader.Download(buf.Bytes())
}

func (c addCommand) play(track string, files []string, replyID int) []messaging.ChatMessage {
	track = strings.ToLower(track)
	for _, f := range files {
		_, name := path.Split(f)
		if strings.Index(strings.ToLower(name), track) >= 0 {
			file, err := c.f.Downloader.GetFile(f)
			if err != nil {
				c.l.Logf(logger.ErrorLevel, "Get file '%s' failed: %s", f, err)
				return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
			}
			msg := messaging.New("Готово", replyID)
			msg.UploadAudio(track, "", file)
			return []messaging.ChatMessage{msg}
		}
	}
	return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
}
