package messaging

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

func (a *channelAdapter) getMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
	return a.channelService.GetMentions(ctx, a.teamRef, ref, rawMentions)
}

func (a *channelAdapter) sendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
	return a.channelService.SendMessage(ctx, a.teamRef, ref, body)
}
