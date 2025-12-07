package utils

import (
	"time"

	"github.com/pzsp-teams/lib/channels"
)

// TimeRange will be used later
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// DisplayMessageInfo will be used later
type DisplayMessageInfo struct {
	TeamName    string
	ChannelName string
	Message     *channels.Message
}
