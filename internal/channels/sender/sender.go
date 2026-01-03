package sender

import (
	"context"

	coresender "github.com/pzsp-teams/cli/internal/core/sender"
	"github.com/pzsp-teams/lib/channels"
)

type channelSender struct {
	adapter *channelAdapter
	generic *coresender.GenericSender[ChannelSendResult]
}

// ChannelSender defines the interface for sending messages to channels
type ChannelSender interface {
	Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []ChannelSendResult
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
