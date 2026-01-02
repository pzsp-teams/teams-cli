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

func (a *channelAdapter) newErrorResult(ref, content string, err error) ChannelSendResult {
	return ChannelSendResult{
		ChannelRef: ref,
		Message:    content,
		Error:      err,
	}
}

func (a *channelAdapter) newSuccessResult(ref, message string) ChannelSendResult {
	return ChannelSendResult{
		ChannelRef: ref,
		Message:    message,
		Error:      nil,
	}
}

func (a *channelAdapter) getError(result ChannelSendResult) error {
	return result.Error
}

func (a *channelAdapter) getLogFields(ref string) map[string]any {
	return map[string]any{
		"team":    a.teamRef,
		"channel": ref,
	}
}
