package play

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

type playCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
}

const (
	maxResult     int64 = 10
	searchTimeout       = 1 * time.Minute
)

var Command command.Type = command.Type{
	ID:       "play",
	Title:    "Прослушать трек",
	Help:     "",
	Factory:  New,
	Internal: true,
}

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return playCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "play"}),
	}
}

func (c playCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	if len(arguments) != 1 {
		return messaging.NewSingleMessage(command.ParseArgumentsFailed, replyID)
	}

	args, ok := registry.Get[*command.DownloadArguments](c.interlayer.Registry, arguments[0])
	if !ok || !args.IsValid() {
		c.l.Logf(logger.WarnLevel, "Possible short link has expired")
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	q := args.Artist

	const token = ""
	cli, auth := c.interlayer.Discovery.New(token)
	var variants []*models.SearchTorrentsResult
	var err error

	variants, err = c.searchTorrents(cli, auth, q)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Search torrents failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	sel := selector.MusicSelector{
		Query:               q,
		Discography:         true,
		MinSeedersThreshold: 10,
		MaxSizeMB:           50 * 1024,
		Format:              "mp3",
	}
	chosen := sel.Select(variants)
	if chosen == nil {
		return messaging.NewSingleMessage(command.NothingFound, replyID)
	}

	files, err := c.download(cli, auth, *chosen.Link)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Enqueue downloading failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	return c.play(args.Track, files, replyID)
}

func (c playCommand) searchTorrents(cli *client.Client, auth runtime.ClientAuthInfoWriter, q string) ([]*models.SearchTorrentsResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	req := torrents.SearchTorrentsParams{
		Limit:       utils.ToPointer(maxResult),
		Q:           q,
		Strong:      utils.ToPointer(false),
		Type:        utils.ToPointer("music"),
		Discography: utils.ToPointer(true),
		Context:     ctx,
	}

	resp, err := cli.Torrents.SearchTorrents(&req, auth)
	if err != nil {
		return nil, err
	}
	return resp.Payload.Results, nil
}

func (c playCommand) download(cli *client.Client, auth runtime.ClientAuthInfoWriter, link string) ([]string, error) {
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

	return c.interlayer.Downloader.Download(buf.Bytes())
}

func (c playCommand) play(track string, files []string, replyID int) []messaging.ChatMessage {
	track = strings.ToLower(track)
	for _, f := range files {
		_, name := path.Split(f)
		if strings.Index(strings.ToLower(name), track) >= 0 {
			file, err := c.interlayer.Downloader.GetFile(f)
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
