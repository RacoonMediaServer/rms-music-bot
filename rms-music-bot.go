package main

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-music-bot/internal/bot"
	"github.com/RacoonMediaServer/rms-music-bot/internal/config"
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/db"
	"github.com/RacoonMediaServer/rms-music-bot/internal/downloader"
	"github.com/RacoonMediaServer/rms-music-bot/internal/service"
	"github.com/go-co-op/gocron"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	f := connectivity.New(conf.Remote, microService)
	f.Downloader = dw

	tgBot, err := bot.New(conf.Bot.Token, service.New(f))
	if err != nil {
		logger.Fatalf("Start bot failed: %s", err)
	}

	if err = dw.Start(); err != nil {
		logger.Fatalf("Start downloader failed: %s", err)
	}

	// clean downloaded data every day at night
	s := gocron.NewScheduler(time.Local)
	_, _ = s.Every(1).Day().At("00:03").Do(func() {
		dw.Wipe()
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	tgBot.Stop()
	dw.Stop()
}
