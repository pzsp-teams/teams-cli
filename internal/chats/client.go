package chats

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/chats/retriever"
	"github.com/pzsp-teams/teams-cli/internal/chats/sender"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
)

// ChatClient provides all chat-related operations
type ChatClient interface {
	// Send sends messages to chats
	Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []sender.ChatSendResult
	// GetMessages retrieves messages from all chats within the specified time range
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, chatRef chats.ChatRef) ([]*retriever.ChatMessageWithContext, error)
	// List returns a list of chats
	List(ctx context.Context) ([]*models.Chat, error)
}

type client struct {
	sender       sender.ChatSender
	retriever    retriever.Retriever
	chatsService chats.Service
}

// NewChatClient creates a new chats client
func NewChatClient(chatsService chats.Service) ChatClient {
	return &client{
		sender:       sender.NewChatSender(chatsService),
		retriever:    retriever.NewRetriever(chatsService),
		chatsService: chatsService,
	}
}

// Send sends messages to chats
func (c *client) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []sender.ChatSendResult {
	return c.sender.Send(ctx, messages, dryRun, ignoreError)
}

// GetMessages retrieves messages from all chats within the specified time range
func (c *client) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, chatRef chats.ChatRef) ([]*retriever.ChatMessageWithContext, error) {
	return c.retriever.GetMessages(ctx, timeRange, chatRef)
}

// List returns a list of chats
func (c *client) List(ctx context.Context) ([]*models.Chat, error) {
	return c.chatsService.ListChats(ctx, nil)
}
