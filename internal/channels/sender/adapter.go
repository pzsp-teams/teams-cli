package sender

import (
	"context"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
)

type channelAdapter struct {
	channelService channels.Service
	teamRef        string
}

func (a *channelAdapter) setTeamRef(teamRef string) {
	a.teamRef = teamRef
}

// GetMentions implements the coresender.Adapter interface
func (a *channelAdapter) GetMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
	return a.channelService.GetMentions(ctx, a.teamRef, ref, rawMentions)
}

// SendMessage implements the coresender.Adapter interface
func (a *channelAdapter) SendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
	return a.channelService.SendMessage(ctx, a.teamRef, ref, body)
}
