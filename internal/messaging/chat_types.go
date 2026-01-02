package messaging

// ChatSendResult represents the result of sending a message to a chat
type ChatSendResult struct {
	ChatRef string
	Message string
	Error   error
}
