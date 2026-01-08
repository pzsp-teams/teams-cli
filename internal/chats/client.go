package chats

import (
	"context"

	"github.com/pzsp-teams/cli/internal/chats/retriever"
	"github.com/pzsp-teams/cli/internal/chats/sender"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

// Client provides all chat-related operations
type Client struct {
	sender       sender.ChatSender
	retriever    retriever.Retriever
	chatsService chats.Service
}

// NewClient creates a new chats client
func NewClient(chatsService chats.Service) *Client {
	return &Client{
		sender:       sender.NewChatSender(chatsService),
		retriever:    retriever.NewRetriever(chatsService),
		chatsService: chatsService,
	}
}

// Send sends messages to chats
func (c *Client) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []sender.ChatSendResult {
	return c.sender.Send(ctx, messages, dryRun, ignoreError)
}

// GetMessages retrieves messages from all chats within the specified time range
func (c *Client) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*retriever.ChatMessageWithContext, error) {
	return c.retriever.GetMessages(ctx, timeRange)
}

// List returns a list of chats
func (c *Client) List(ctx context.Context) ([]*models.Chat, error) {
	return c.chatsService.ListChats(ctx, nil)
}
