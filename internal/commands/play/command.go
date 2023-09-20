package play

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"github.com/RacoonMediaServer/rms-music-bot/internal/provider"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-music-bot/internal/selector"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"github.com/go-openapi/runtime"
	"go-micro.dev/v4/logger"
	"time"
)

type playCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
}

var errNotFound = errors.New("nothing found")

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
	directory, err := c.findExistingDirectory(args)
	if err != nil {
		directory, err = c.download(args, model.Discography)
		if err != nil {
			if errors.Is(err, errNotFound) {
				return messaging.NewSingleMessage(command.NothingFound, replyID)
			}
			c.l.Logf(logger.ErrorLevel, "Download content failed: %s", err)
			return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
		}
	}

	return c.play(args, directory, replyID)
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

func (c playCommand) getTorrentFile(cli *client.Client, auth runtime.ClientAuthInfoWriter, link string) ([]byte, error) {
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

	return buf.Bytes(), nil
}

func (c playCommand) download(args *command.DownloadArguments, contentType model.ContentType) (string, error) {
	var token = config.Config().Token // TODO: remove
	cli, auth := c.interlayer.Discovery.New(token)
	var variants []*models.SearchTorrentsResult
	var err error

	title := args.Artist
	if contentType == model.Album {
		title = args.Album
	}

	variants, err = c.searchTorrents(cli, auth, title)
	if err != nil {
		return "", fmt.Errorf("search torrents failed: %w", err)
	}

	sel := selector.MusicSelector{
		Query:               title,
		Discography:         true,
		MinSeedersThreshold: 10,
		MaxSizeMB:           50 * 1024,
		Format:              "mp3",
	}
	if contentType == model.Album {
		sel.Discography = false
	}
	chosen := sel.Select(variants)
	if chosen == nil {
		return "", errNotFound
	}

	torrentFile, err := c.getTorrentFile(cli, auth, *chosen.Link)
	if err != nil {
		return "", fmt.Errorf("get torrent file failed: %w", err)
	}

	directory, err := c.interlayer.TorrentManager.Add(torrentFile)
	if err != nil {
		return "", fmt.Errorf("enqueue downloading failed: %w", err)
	}

	contentItem := model.Content{
		Title: title,
		Type:  contentType,
		Torrent: model.Torrent{
			Title: directory,
			Bytes: torrentFile,
		},
	}

	if err = c.interlayer.ContentManager.AddContent(args.Artist, contentItem); err != nil {
		c.l.Logf(logger.WarnLevel, "Save downloading to persistent storage failed: %s", err)
	}

	return directory, nil
}

func (c playCommand) findExistingDirectory(args *command.DownloadArguments) (string, error) {
	content, err := c.interlayer.ContentManager.GetContent(args.Artist)
	if err != nil {
		return "", err
	}

	for _, item := range content {
		if item.Type != model.Album {
			continue
		}
		if item.Title == args.Album {
			return item.Torrent.Title, nil
		}
	}

	for _, item := range content {
		if item.Type != model.Discography {
			continue
		}
		return item.Torrent.Title, nil
	}
	return "", errNotFound
}

func (c playCommand) play(args *command.DownloadArguments, contentDir string, replyID int) []messaging.ChatMessage {
	data, err := c.interlayer.ContentProvider.FindTrack(contentDir, args.Track)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Find track '%s' failed: %s", args.Track, err)
		if errors.Is(err, provider.ErrNotFound) {
			return messaging.NewSingleMessage(command.NothingFound, replyID)
		}
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	title := fmt.Sprintf("%s - %s", args.Artist, args.Track)
	msg := messaging.New("Готово", replyID)
	msg.UploadAudio(title, "", data)
	return []messaging.ChatMessage{msg}
}
