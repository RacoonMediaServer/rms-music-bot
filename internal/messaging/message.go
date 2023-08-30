package messaging

import (
	"bytes"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
)

type ChatMessage interface {
	Compose(chatID int64) tgbotapi.Chattable
}

type Message struct {
	text             string
	replyToMessageID int
	keyboard         KeyboardStyle
	attachment       interface{}

	buttons []button
}

type videoUploading struct {
	name     string
	mimeType string
	rd       io.Reader
}

type audioUploading struct {
	name string
	rd   io.Reader
}

type photoUploading struct {
	name     string
	mimeType string
	image    []byte
}

type photoURL string

func New(text string, replyToMessageID int) *Message {
	return &Message{
		text:             text,
		replyToMessageID: replyToMessageID,
	}
}

func NewSingleMessage(text string, replyToMessageID int) []ChatMessage {
	return []ChatMessage{New(text, replyToMessageID)}
}

func (m *Message) AddButton(title string, command string) {
	m.buttons = append(m.buttons, button{
		command: command,
		title:   title,
	})
}

func (m *Message) SetPhotoURL(url string) {
	u := photoURL(url)
	m.attachment = &u
}

func (m *Message) UploadPhoto(name, mimeType string, image []byte) {
	m.attachment = &photoUploading{
		name:     name,
		mimeType: mimeType,
		image:    image,
	}
}

func (m *Message) UploadAudio(name, mimeType string, audio []byte) {
	m.attachment = &audioUploading{
		name: name,
		rd:   bytes.NewReader(audio),
	}
}

func (m *Message) SetKeyboardStyle(style KeyboardStyle) {
	m.keyboard = style
}

func (m *Message) UploadVideo(name, mimeType string, video []byte) {
	m.attachment = &videoUploading{
		name:     name,
		mimeType: mimeType,
		rd:       bytes.NewReader(video),
	}
}

func (m *Message) Compose(chatID int64) tgbotapi.Chattable {
	var msg tgbotapi.Chattable
	var keyboard interface{}

	if len(m.buttons) > 0 {
		if m.keyboard == ChatKeyboard {
			buttons := make([]tgbotapi.KeyboardButton, 0)
			for _, button := range m.buttons {
				buttons = append(buttons, tgbotapi.NewKeyboardButton(button.command))
			}

			keyboard = tgbotapi.NewReplyKeyboard(buttons)
		} else {
			buttons := make([]tgbotapi.InlineKeyboardButton, 0)
			for _, action := range m.buttons {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(action.title, action.command))
			}

			keyboard = tgbotapi.NewInlineKeyboardMarkup(buttons)
		}
	}

	if m.attachment != nil {
		switch attach := m.attachment.(type) {
		case *photoURL:
			photoMessage := tgbotapi.PhotoConfig{}
			photoMessage.Caption = m.text
			photoMessage.FileID = string(*attach)
			photoMessage.ParseMode = "HTML"
			photoMessage.UseExisting = true
			photoMessage.ReplyMarkup = keyboard
			photoMessage.ChatID = chatID
			photoMessage.ReplyToMessageID = m.replyToMessageID

			msg = photoMessage

		case *photoUploading:
			fileBytes := tgbotapi.FileBytes{
				Name:  attach.name,
				Bytes: attach.image,
			}

			photoMessage := tgbotapi.NewPhotoUpload(chatID, fileBytes)
			photoMessage.MimeType = attach.mimeType
			photoMessage.Caption = m.text
			photoMessage.ParseMode = "HTML"
			photoMessage.ReplyMarkup = keyboard
			photoMessage.ReplyToMessageID = m.replyToMessageID

			msg = photoMessage

		case *videoUploading:
			fileReader := tgbotapi.FileReader{
				Name:   attach.name,
				Size:   -1,
				Reader: attach.rd,
			}
			videoMessage := tgbotapi.NewVideoUpload(chatID, fileReader)
			videoMessage.MimeType = attach.mimeType
			videoMessage.Caption = m.text
			videoMessage.ParseMode = "HTML"
			videoMessage.ReplyMarkup = keyboard
			videoMessage.ReplyToMessageID = m.replyToMessageID

			msg = videoMessage

		case *audioUploading:
			fileReader := tgbotapi.FileReader{
				Name:   attach.name,
				Size:   -1,
				Reader: attach.rd,
			}
			audioMessage := tgbotapi.NewAudioUpload(chatID, fileReader)
			audioMessage.Caption = m.text
			audioMessage.ParseMode = "HTML"
			audioMessage.ReplyMarkup = keyboard
			audioMessage.ReplyToMessageID = m.replyToMessageID

			msg = audioMessage
		}
	} else {
		textMessage := tgbotapi.NewMessage(chatID, m.text)
		textMessage.ParseMode = "HTML"
		textMessage.ReplyToMessageID = m.replyToMessageID
		textMessage.ReplyMarkup = keyboard

		msg = &textMessage
	}

	return msg
}
