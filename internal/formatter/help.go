package formatter

import (
	"bytes"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
)

type HelpContext struct {
	Link     string
	UserName string
	Password string
	Text     string
}

func FormatHelp(ctx HelpContext, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "help", ctx); err != nil {
		logger.Logf(logger.ErrorLevel, "execute template failed: %s", err)
	}

	return messaging.New(buf.String(), replyID)
}
