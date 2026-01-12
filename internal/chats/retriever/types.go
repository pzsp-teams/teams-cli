package retriever

import (
	"context"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

// ChatMessageWithContext contains a message with its chat context
type ChatMessageWithContext struct {
	ChatName string
	ChatID   string
	ChatType string
	Message  *models.Message
}

// Retriever defines the interface for retrieving chat messages
type Retriever interface {
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, chatRef chats.ChatRef) ([]*ChatMessageWithContext, error)
}
