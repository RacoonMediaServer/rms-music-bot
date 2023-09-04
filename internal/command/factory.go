package command

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/connectivity"
	"go-micro.dev/v4/logger"
)

// Factory can create Command of specified type. Factory knows all about specific command
type Factory func(interlayer connectivity.Interlayer, l logger.Logger) Command
