package retriever

import (
	"context"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/models"
)

// Retriever defines the interface for retrieving chat messages
type Retriever interface {
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*models.Message, error)
}
