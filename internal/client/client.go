package client

import (
	"context"

	channelcreation "github.com/pzsp-teams/cli/internal/channel_creation"
	"github.com/pzsp-teams/cli/internal/messaging"
	"github.com/pzsp-teams/lib"
)

// TeamsClient aggregates service wrappers for the Teams API
type TeamsClient struct {
	ChannelSender  messaging.Sender
	ChannelCreator channelcreation.ChannelCreator
}

// NewTeamsClient creates a new TeamsClient by constructing the underlying library client
func NewTeamsClient(ctx context.Context, authConfig *lib.AuthConfig, senderConfig *lib.SenderConfig) (*TeamsClient, error) {
	libClient, err := lib.NewClient(ctx, authConfig, senderConfig, true, nil)
	if err != nil {
		return nil, err
	}

	return &TeamsClient{
		ChannelSender:  messaging.NewChannelSender(libClient.Channels),
		ChannelCreator: channelcreation.NewChannelCreator(libClient.Channels),
	}, nil
}
