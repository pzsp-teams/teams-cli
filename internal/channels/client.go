package channels

import (
	"context"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/teams"
	"github.com/pzsp-teams/teams-cli/internal/channels/creator"
	"github.com/pzsp-teams/teams-cli/internal/channels/retriever"
	"github.com/pzsp-teams/teams-cli/internal/channels/sender"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
)

// ChannelClient provides all channel-related operations
type ChannelClient interface {
	// Send sends messages to channels
	Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []sender.ChannelSendResult
	// SendReply sends a reply to a specific message in a channel
	SendReply(ctx context.Context, teamRef, channelRef, messageID, content string) sender.ChannelSendResult
	// Create creates channels based on the provided request data
	Create(ctx context.Context, teamRef string, request map[string]creator.ChannelData, ensureMembersInChannel, ensureMembersInTeam, dryRun bool) []creator.CreateResult
	// GetMessages retrieves messages from all channels within the specified time range
	GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*retriever.ChannelMessageWithContext, error)
}

// NewChannelClient creates a new channels client
func NewChannelClient(channelsService channels.Service, teamsService teams.Service) ChannelClient {
	return &client{
		sender:    sender.NewChannelSender(channelsService),
		creator:   creator.NewChannelCreator(channelsService, teamsService),
		retriever: retriever.NewRetriever(teamsService, channelsService),
	}
}

type client struct {
	sender    sender.ChannelSender
	creator   creator.ChannelCreator
	retriever retriever.Retriever
}

// Send sends messages to channels
func (c *client) Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []sender.ChannelSendResult {
	return c.sender.Send(ctx, teamRef, messages, dryRun, ignoreError)
}

// SendReply sends a reply to a specific message in a channel
func (c *client) SendReply(ctx context.Context, teamRef, channelRef, messageID, content string) sender.ChannelSendResult {
	return c.sender.SendReply(ctx, teamRef, channelRef, messageID, content)
}

// Create creates channels based on the provided request data
func (c *client) Create(ctx context.Context, teamRef string, request map[string]creator.ChannelData, ensureMembersInChannel, ensureMembersInTeam, dryRun bool) []creator.CreateResult {
	return c.creator.Create(ctx, teamRef, request, ensureMembersInChannel, ensureMembersInTeam, dryRun)
}

// GetMessages retrieves messages from all channels within the specified time range
func (c *client) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*retriever.ChannelMessageWithContext, error) {
	return c.retriever.GetMessages(ctx, timeRange, teamRef, channelRef)
}
