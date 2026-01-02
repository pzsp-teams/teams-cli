package messaging

// ChannelSendResult represents the result of sending a message to a channel
type ChannelSendResult struct {
	ChannelRef string
	Message    string
	Error      error
}
