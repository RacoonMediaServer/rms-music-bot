package add

import (
	"bytes"
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/model"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/RacoonMediaServer/rms-music-bot/internal/selector"
	"github.com/RacoonMediaServer/rms-music-bot/internal/utils"
	"github.com/go-openapi/runtime"
	"go-micro.dev/v4/logger"
	"time"
)

type addCommand struct {
	interlayer connectivity.Interlayer
	l          logger.Logger
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

func New(interlayer connectivity.Interlayer, l logger.Logger) command.Command {
	return addCommand{
		interlayer: interlayer,
		l:          l.Fields(map[string]interface{}{"command": "add"}),
	}
}

func (c addCommand) Do(arguments command.Arguments, replyID int) []messaging.ChatMessage {
	if len(arguments) != 1 {
		return messaging.NewSingleMessage(command.ParseArgumentsFailed, replyID)
	}

	args, ok := registry.Get[*command.DownloadArguments](c.interlayer.Registry, arguments[0])
	if !ok || !args.IsValid() {
		c.l.Logf(logger.WarnLevel, "Possible short link has expired")
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	q := args.Artist
	if args.Album != "" {
		q += " " + args.Album
	}

	allAlbums := args.Album == "" && args.Track == ""

	var token = config.Config().Token // TODO: remove
	cli, auth := c.interlayer.Discovery.New(token)

	variants, err := c.searchTorrents(cli, auth, q, allAlbums)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Search torrents failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}
	if len(variants) == 0 {
		return messaging.NewSingleMessage(command.NothingFound, replyID)
	}

	sel := selector.MusicSelector{
		Query:               q,
		Discography:         allAlbums,
		MinSeedersThreshold: 10,
		MaxSizeMB:           50 * 1024,
		Format:              "mp3",
	}
	chosen := sel.Select(variants)

	torrentFile, err := c.getTorrentFile(cli, auth, *chosen.Link)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Get torrent file failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	directory, err := c.interlayer.TorrentManager.Add(torrentFile)
	if err != nil {
		c.l.Logf(logger.ErrorLevel, "Enqueue downloading failed: %s", err)
		return messaging.NewSingleMessage(command.SomethingWentWrong, replyID)
	}

	contentItem := model.Content{
		Title: args.Album,
		Torrent: model.Torrent{
			Title: directory,
			Bytes: torrentFile,
		},
	}

	if !allAlbums {
		contentItem.Type = model.Album
	}

	if err = c.interlayer.ContentManager.AddContent(args.Artist, contentItem); err != nil {
		c.l.Logf(logger.WarnLevel, "Save downloading to persistent storage failed: %s", err)
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

func (c addCommand) getTorrentFile(cli *client.Client, auth runtime.ClientAuthInfoWriter, link string) ([]byte, error) {
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
