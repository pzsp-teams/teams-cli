package retriever

import (
	"context"

	"github.com/pzsp-teams/lib/models"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
)

// ChannelMessageWithContext contains a message with its team and channel context
type ChannelMessageWithContext struct {
	TeamName    string
	TeamID      string
	ChannelName string
	ChannelID   string
	Message     *models.Message
}

// Retriever defines the interface for retrieving channel messages
type Retriever interface {
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*ChannelMessageWithContext, error)
}
