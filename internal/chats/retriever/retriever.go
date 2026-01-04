package retriever

import (
	"context"
	"fmt"

	f "github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

type retriever struct {
	chatService chats.Service
	formatter   f.Formatter
}

// NewRetriever creates a new chat message retriever
func NewRetriever(chatService chats.Service, formatter f.Formatter) Retriever {
	return &retriever{
		chatService: chatService,
		formatter:   formatter,
	}
}

// GetMessages retrieves messages from all chats within the specified time range
func (r *retriever) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*models.Message, error) {
	top := int32(50) // Default pagination size

	messages, err := r.chatService.ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListingMessagesFailed, err)
	}

	for _, msg := range messages {
		if msg.ContentType == models.MessageContentTypeHTML {
			msg.Content = r.formatter.Format(msg.Content)
		}
	}

	return messages, nil
}
