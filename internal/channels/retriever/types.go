package retriever

import (
	"time"

	"github.com/pzsp-teams/lib/models"
)

// TimeRange defines a time window for message retrieval
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// DisplayMessageInfo contains information about a message including its context
type DisplayMessageInfo struct {
	TeamName    string
	ChannelName string
	Message     *models.Message
}
