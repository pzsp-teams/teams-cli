package sender

// ChannelSendResult represents the result of sending a message to a channel
type ChannelSendResult struct {
	ChannelRef string
	Message    string
	Error      error
}

// GetError implements the coresender.Result interface
func (r ChannelSendResult) GetError() error {
	return r.Error
}

// GetRef implements the coresender.Result interface
func (r ChannelSendResult) GetRef() string {
	return r.ChannelRef
}

// GetMessage implements the coresender.Result interface
func (r ChannelSendResult) GetMessage() string {
	return r.Message
}
