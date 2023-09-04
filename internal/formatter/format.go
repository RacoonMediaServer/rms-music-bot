package formatter

import (
	"bytes"
	"embed"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-music-bot/internal/command"
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"go-micro.dev/v4/logger"
	"text/template"
	"time"
)

//go:embed templates
var templates embed.FS

const uidLifeTime = 2 * time.Hour

var parsedTemplates *template.Template

func init() {
	parsedTemplates = template.Must(template.ParseFS(templates, "templates/*.txt"))
}

func (f formatter) formatArtist(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "artist", r); err != nil {
		f.l.Logf(logger.ErrorLevel, "Execute formatter template failed: %s", err)
	}

	downloadArgs := command.DownloadArguments{
		Artist: r.Artist,
	}
	uid := f.r.Add(&downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Добавить дискографию", "/add "+uid)
	return m
}

func (f formatter) formatAlbum(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "album", r); err != nil {
		f.l.Logf(logger.ErrorLevel, "Execute formatter template failed: %s", err)
	}

	downloadArgs := command.DownloadArguments{
		Artist: r.Artist,
		Album:  r.Album,
	}
	uid := f.r.Add(&downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Добавить альбом", "/add "+uid)

	return m
}

func (f formatter) formatTrack(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "track", r); err != nil {
		f.l.Logf(logger.ErrorLevel, "execute template failed: %s", err)
	}

	downloadArgs := command.DownloadArguments{
		Artist: r.Artist,
		Album:  r.Album,
		Track:  *r.Title,
	}
	uid := f.r.Add(&downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Слушать трек", "/play "+uid)
	return m
}

func (f formatter) FormatSearchMusicResult(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	if r.Type == nil {
		f.l.Logf(logger.ErrorLevel, "unknown message type")
		return nil
	}

	switch *r.Type {
	case "artist":
		return f.formatArtist(r, replyID)
	case "album":
		return f.formatAlbum(r, replyID)
	case "track":
		return f.formatTrack(r, replyID)
	}

	f.l.Logf(logger.ErrorLevel, "unknown message type")
	return nil
}
