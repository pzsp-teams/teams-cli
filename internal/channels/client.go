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

// Client provides all channel-related operations
type Client struct {
	sender    sender.ChannelSender
	creator   creator.ChannelCreator
	retriever retriever.Retriever
}

// NewClient creates a new channels client
func NewClient(channelsService channels.Service, teamsService teams.Service) *Client {
	return &Client{
		sender:    sender.NewChannelSender(channelsService),
		creator:   creator.NewChannelCreator(channelsService, teamsService),
		retriever: retriever.NewRetriever(teamsService, channelsService),
	}
}

// Send sends messages to channels
func (c *Client) Send(ctx context.Context, teamRef string, messages map[string]string, dryRun, ignoreError bool) []sender.ChannelSendResult {
	return c.sender.Send(ctx, teamRef, messages, dryRun, ignoreError)
}

// SendReply sends a reply to a specific message in a channel
func (c *Client) SendReply(ctx context.Context, teamRef, channelRef, messageID, content string) sender.ChannelSendResult {
	return c.sender.SendReply(ctx, teamRef, channelRef, messageID, content)
}

// Create creates channels based on the provided request data
func (c *Client) Create(ctx context.Context, teamRef string, request map[string]creator.ChannelData, ensureMembersInChannel, ensureMembersInTeam, dryRun bool) []creator.CreateResult {
	return c.creator.Create(ctx, teamRef, request, ensureMembersInChannel, ensureMembersInTeam, dryRun)
}

// GetMessages retrieves messages from all channels within the specified time range
func (c *Client) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, teamRef, channelRef *string) ([]*retriever.ChannelMessageWithContext, error) {
	return c.retriever.GetMessages(ctx, timeRange, teamRef, channelRef)
}
