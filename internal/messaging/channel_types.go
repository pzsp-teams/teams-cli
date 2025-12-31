package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

// SendResult represents the result of sending a message to a channel
type SendResult struct {
	ChannelRef string
	Message    string
	Error      error
}

type sendMessageData struct {
	teamRef    string
	channelRef string
	body       models.MessageBody
}

type action struct {
	*sendMessageData
	run    func(ctx context.Context, message *sendMessageData) *SendResult
	result *SendResult
}
