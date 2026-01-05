package retriever

import (
	"github.com/pzsp-teams/lib/models"
)

// ChannelMessageWithContext contains a message with its team and channel context
type ChannelMessageWithContext struct {
	TeamName    string
	TeamID      string
	ChannelName string
	ChannelID   string
	Message     *models.Message
}
