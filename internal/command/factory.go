package command

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"go-micro.dev/v4/logger"
)

// Factory can create Command of specified type. Factory knows all about specific command
type Factory func(f connectivity.Interlayer, l logger.Logger, r registry.Registry) Command
