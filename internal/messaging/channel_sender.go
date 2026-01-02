package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/channels"
)

type channelSender struct {
	generic *genericSender[ChannelSendResult]
}

// ChannelSender defines the interface for sending messages to channels
type ChannelSender interface {
	Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []ChannelSendResult
}

// NewChannelSender creates a new ChannelSender with the provided channels service
func NewChannelSender(channelService channels.Service) *channelSender {
	adapter := &channelAdapter{
		channelService: channelService,
	}
	return &channelSender{
		generic: newGenericSender(adapter),
	}
}

// Send sends messages to multiple channels within a team
//
// teamRef: team name or ID
// messages: map of channel reference (name or ID) to message content
// Returns a slice of SendResult containing the outcome for each channel
func (s *channelSender) Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []ChannelSendResult {
	s.generic.adapter.(*channelAdapter).teamRef = teamRef
	return s.generic.send(ctx, messages, dryRun, ignoreError)
}
