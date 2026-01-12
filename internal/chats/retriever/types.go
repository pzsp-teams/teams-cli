package retriever

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
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
