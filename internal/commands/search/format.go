package search

import (
	"bytes"
	"embed"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
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

func (s searchCommand) formatResult(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	if r.Type == nil {
		return nil
	}
	switch *r.Type {
	case "artist":
		return s.formatArtist(r, replyID)
	case "album":
		return s.formatAlbum(r, replyID)
	case "track":
		return s.formatTrack(r, replyID)
	default:
		return nil
	}
}

func (s searchCommand) formatArtist(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "artist", r); err != nil {
		s.l.Logf(logger.ErrorLevel, "execute template failed: %s", err)
	}

	downloadArgs := map[string]string{
		"artist": r.Artist,
	}
	uid := s.r.Add(downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Добавить дискографию", "/add "+uid)
	return m
}

func (s searchCommand) formatAlbum(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "album", r); err != nil {
		s.l.Logf(logger.ErrorLevel, "execute template failed: %s", err)
	}

	downloadArgs := map[string]string{
		"artist": r.Artist,
		"album":  r.Album,
	}
	uid := s.r.Add(downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Добавить альбом", "/add "+uid)

	return m
}

func (s searchCommand) formatTrack(r *models.SearchMusicResult, replyID int) messaging.ChatMessage {
	var buf bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buf, "track", r); err != nil {
		s.l.Logf(logger.ErrorLevel, "execute template failed: %s", err)
	}

	downloadArgs := map[string]string{
		"artist": r.Artist,
		"album":  r.Album,
		"track":  *r.Title,
	}
	uid := s.r.Add(downloadArgs, uidLifeTime)

	m := messaging.New(buf.String(), replyID)
	m.SetPhotoURL(r.Picture)
	m.SetKeyboardStyle(messaging.MessageKeyboard)
	m.AddButton("Добавить трек", "/add "+uid)
	m.AddButton("Слушать трек", "/play "+uid)
	return m
}
