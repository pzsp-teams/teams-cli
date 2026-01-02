package messaging

// ChatSendResult represents the result of sending a message to a chat
type ChatSendResult struct {
	ChatRef string
	Message string
	Error   error
}

func (r ChatSendResult) getError() error {
	return r.Error
}

func (r ChatSendResult) getRef() string {
	return r.ChatRef
}

func (r ChatSendResult) getMessage() string {
	return r.Message
}
