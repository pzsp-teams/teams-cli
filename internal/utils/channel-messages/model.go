package utils

import (
	"time"

	"github.com/pzsp-teams/lib/channels"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type DisplayMessageInfo struct {
	TeamName    string
	ChannelName string
	Message     *channels.Message
}
