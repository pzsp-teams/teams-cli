package client

import (
	"context"

	"github.com/pzsp-teams/lib"
	lib_config "github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/teams-cli/internal/channels"
	"github.com/pzsp-teams/teams-cli/internal/chats"
	"github.com/pzsp-teams/teams-cli/internal/teams"
)

// Client aggregates resource-level clients
type Client struct {
	Channels channels.ChannelClient
	Chats    chats.ChatClient
	Teams    teams.TeamClient
}

// NewClient creates a new TeamsClient by constructing the underlying library client
func NewClient(ctx context.Context, authConfig *lib_config.AuthConfig, senderConfig *lib_config.SenderConfig, cacheConfig *lib_config.CacheConfig) (*Client, error) {
	libClient, err := lib.NewClient(ctx, authConfig, senderConfig, cacheConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		Channels: channels.NewChannelClient(libClient.Channels, libClient.Teams),
		Chats:    chats.NewChatClient(libClient.Chats),
		Teams:    teams.NewTeamClient(libClient.Teams),
	}, nil
}
