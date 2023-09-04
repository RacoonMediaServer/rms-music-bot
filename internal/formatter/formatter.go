package formatter

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"github.com/RacoonMediaServer/rms-music-bot/internal/registry"
	"go-micro.dev/v4/logger"
)

type Formatter interface {
	FormatSearchMusicResult(r *models.SearchMusicResult, replyID int) messaging.ChatMessage
}

type formatter struct {
	l logger.Logger
	r registry.Registry
}

func New(l logger.Logger, r registry.Registry) Formatter {
	return &formatter{
		l: l,
		r: r,
	}
}
