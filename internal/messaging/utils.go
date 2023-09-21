package messaging

func Reverse(messages []ChatMessage) []ChatMessage {
	result := make([]ChatMessage, len(messages))
	for i := range messages {
		result[len(messages)-i-1] = messages[i]
	}
	return result
}
