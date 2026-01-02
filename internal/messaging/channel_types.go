package messaging

// ChannelSendResult represents the result of sending a message to a channel
type ChannelSendResult struct {
	ChannelRef string
	Message    string
	Error      error
}

func (r ChannelSendResult) getError() error {
	return r.Error
}

func (r ChannelSendResult) getRef() string {
	return r.ChannelRef
}

func (r ChannelSendResult) getMessage() string {
	return r.Message
}
