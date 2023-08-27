package bot

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	incomingMessagesCounter *prometheus.CounterVec
)

func init() {
	incomingMessagesCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "rms",
		Subsystem: "music_bot",
		Name:      "incoming_messages_count",
		Help:      "Total count of messages from user",
	}, []string{"user"})
}
