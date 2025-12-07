package utils

import (
	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/teams"
)

type DisplayMessageInfo struct {
	Team         *teams.Team
	Channel      *channels.Channel
	Message 	 *channels.Message
}