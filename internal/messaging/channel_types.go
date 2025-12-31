package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

// ChannelSendResult represents the result of sending a message to a channel
type ChannelSendResult struct {
	ChannelRef string
	Message    string
	Error      error
}

type channelMessageData struct {
	teamRef    string
	channelRef string
	body       models.MessageBody
}

type channelAction struct {
	*channelMessageData
	run    func(ctx context.Context, message *channelMessageData) *ChannelSendResult
	result *ChannelSendResult
}
