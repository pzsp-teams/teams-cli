package retriever

import (
	"github.com/pzsp-teams/lib/models"
)

// DisplayMessageInfo contains information about a message including its context
type DisplayMessageInfo struct {
	TeamName    string
	ChannelName string
	Message     *models.Message
}
