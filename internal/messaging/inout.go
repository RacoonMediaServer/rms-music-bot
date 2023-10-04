package messaging

type Incoming struct {
	ID       int
	ChatID   int64
	UserID   int
	UserName string
	Text     string
}

type Outgoing struct {
	ChatID  int64
	Message ChatMessage
}
