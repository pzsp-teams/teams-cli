package sender

// ChatSendResult represents the result of sending a message to a chat
type ChatSendResult struct {
	ChatRef string
	Message string
	Error   error
}

// GetError implements the coresender.Result interface
func (r ChatSendResult) GetError() error {
	return r.Error
}

// GetRef implements the coresender.Result interface
func (r ChatSendResult) GetRef() string {
	return r.ChatRef
}

// GetMessage implements the coresender.Result interface
func (r ChatSendResult) GetMessage() string {
	return r.Message
}
