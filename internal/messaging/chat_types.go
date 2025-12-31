package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

// ChatSendResult represents the result of sending a message to a chat
type ChatSendResult struct {
	ChatRef string
	Message string
	Error   error
}

type chatMessageData struct {
	chatRef string
	body    models.MessageBody
}

type chatAction struct {
	*chatMessageData
	run    func(ctx context.Context, message *chatMessageData) *ChatSendResult
	result *ChatSendResult
}
