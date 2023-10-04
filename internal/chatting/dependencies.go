package chatting

import "github.com/RacoonMediaServer/rms-music-bot/internal/messaging"

type Messenger interface {
	Incoming() <-chan *messaging.Incoming
	Outgoing() chan<- *messaging.Outgoing
}
