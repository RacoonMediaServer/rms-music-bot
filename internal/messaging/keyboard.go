package messaging

type KeyboardStyle int

const (
	ChatKeyboard KeyboardStyle = iota
	MessageKeyboard
)

type button struct {
	command string
	title   string
}
