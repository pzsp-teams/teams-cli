package sender

import (
	"context"

	coresender "github.com/pzsp-teams/cli/internal/core/sender"
	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
)

type channelSender struct {
	adapter *channelAdapter
	generic *coresender.GenericSender[ChannelSendResult]
	service channels.Service
}

// ChannelSender defines the interface for sending messages to channels
type ChannelSender interface {
	Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []ChannelSendResult
	SendReply(ctx context.Context, teamRef, channelRef, messageID, content string) ChannelSendResult
}

// NewChannelSender creates a new ChannelSender with the provided channels service
func NewChannelSender(channelService channels.Service) ChannelSender {
	adapter := &channelAdapter{
		channelService: channelService,
	}
	newResult := func(ref, message string, err error) ChannelSendResult {
		return ChannelSendResult{ChannelRef: ref, Message: message, Error: err}
	}
	return &channelSender{
		adapter: adapter,
		generic: coresender.NewGenericSender(adapter, newResult, coresender.TypeChannel),
		service: channelService,
	}
}

// Send sends messages to multiple channels within a team
//
// teamRef: team name or ID
// messages: map of channel reference (name or ID) to message content
// Returns a slice of SendResult containing the outcome for each channel
func (s *channelSender) Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []ChannelSendResult {
	s.adapter.setTeamRef(teamRef)
	return s.generic.Send(ctx, messages, dryRun, ignoreError)
}

// Send sends messages to multiple channels within a team
//
// teamRef: team name or ID
// channelRef: channel name or ID
// messageID: ID of message that reply is for
// content: reply content
func (s *channelSender) SendReply(ctx context.Context, teamRef, channelRef, messageID, content string) ChannelSendResult {
	message := models.MessageBody{Content: content, ContentType: models.MessageContentTypeHTML}
	_, err := s.service.SendReply(ctx, teamRef, channelRef, messageID, message)
	return ChannelSendResult{ChannelRef: channelRef, Message: content, Error: err}
}
