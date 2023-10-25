package main

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/access"
	"github.com/RacoonMediaServer/rms-music-bot/internal/bot"
	"github.com/RacoonMediaServer/rms-music-bot/internal/chatting"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/db"
	"github.com/RacoonMediaServer/rms-music-bot/internal/downloader"
	"github.com/RacoonMediaServer/rms-music-bot/internal/health"
	"github.com/RacoonMediaServer/rms-music-bot/internal/provider"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"os"
	"os/signal"
	"syscall"
)

var Version = "v0.0.0"

const serviceName = "rms-music-bot"

func main() {
	logger.Infof("%s %s", serviceName, Version)
	defer logger.Info("DONE.")

	useDebug := false

	microService := micro.NewService(
		micro.Name(serviceName),
		micro.Version(Version),
		micro.Flags(
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"debug"},
				Usage:       "debug log level",
				Value:       false,
				Destination: &useDebug,
			},
		),
	)

	microService.Init(
		micro.Action(func(context *cli.Context) error {
			configFile := fmt.Sprintf("/etc/rms/%s.json", serviceName)
			if context.IsSet("config") {
				configFile = context.String("config")
			}
			return config.Load(configFile)
		}),
	)

	if useDebug {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	conf := config.Config()

	database, err := db.Connect(conf.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}

	dw := downloader.New(conf.Layout, database)
	interlayer := connectivity.New(conf.Remote, microService)
	interlayer.TorrentManager = dw
	interlayer.Registry = registry.New()
	interlayer.ContentManager = database
	interlayer.ContentProvider = provider.NewContentProvider(conf.Layout.Directory)
	interlayer.AccessService = access.New(interlayer.Services, conf)
	interlayer.ChatStorage = database

	tgBot, err := bot.New(conf.Bot.Token)
	if err != nil {
		logger.Fatalf("Start bot failed: %s", err)
	}

	if err = dw.Start(); err != nil {
		logger.Fatalf("Start downloader failed: %s", err)
	}

	chk := health.NewChecker(conf)
	manager := chatting.NewManager(tgBot, interlayer)
	manager.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	logger.Info("Shutdown...")

	chk.Stop()
	manager.Stop()
	tgBot.Stop()
	dw.Stop()
}
