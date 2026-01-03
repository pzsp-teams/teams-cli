package chats

import (
	"context"

	"github.com/pzsp-teams/cli/internal/chats/sender"
	"github.com/pzsp-teams/lib/chats"
)

// Client provides all chat-related operations
type Client struct {
	sender sender.ChatSender
	// TODO: add retriever
}

// NewClient creates a new chats client
func NewClient(chatsService chats.Service) *Client {
	return &Client{
		sender: sender.NewChatSender(chatsService),
	}
}

// Send sends messages to chats
func (c *Client) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []sender.ChatSendResult {
	return c.sender.Send(ctx, messages, dryRun, ignoreError)
}
