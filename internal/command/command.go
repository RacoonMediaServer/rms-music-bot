package command

import (
	"github.com/RacoonMediaServer/rms-music-bot/internal/messaging"
	"strings"
)

// IsCommand checks the text can be interpreted as command
func IsCommand(text string) bool {
	if text == "" {
		return false
	}
	return text[0] == '/'
}

type Command interface {
	Do(ctx Context) []messaging.ChatMessage
}

// Parse splits text string to command name and arguments
func Parse(text string) (command string, arguments Arguments) {
	list := strings.Split(text, " ")
	if len(list) == 0 {
		return
	}
	command = strings.TrimPrefix(list[0], "/")
	arguments = list[1:]
	return
}
