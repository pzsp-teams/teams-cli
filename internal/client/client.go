package client

import (
	"context"

	"github.com/pzsp-teams/cli/internal/channels"
	"github.com/pzsp-teams/cli/internal/chats"
	"github.com/pzsp-teams/cli/internal/teams"
	"github.com/pzsp-teams/lib"
	lib_config "github.com/pzsp-teams/lib/config"
)

// TeamsClient aggregates resource-level clients
type TeamsClient struct {
	Channels *channels.Client
	Chats    *chats.Client
	Teams    *teams.Client
}

// NewTeamsClient creates a new TeamsClient by constructing the underlying library client
func NewTeamsClient(ctx context.Context, authConfig *lib_config.AuthConfig, senderConfig *lib_config.SenderConfig, cacheConfig *lib_config.CacheConfig) (*TeamsClient, error) {
	libClient, err := lib.NewClient(ctx, authConfig, senderConfig, cacheConfig)
	if err != nil {
		return nil, err
	}

	return &TeamsClient{
		Channels: channels.NewClient(libClient.Channels, libClient.Teams),
		Chats:    chats.NewClient(libClient.Chats),
		Teams:    teams.NewClient(libClient.Teams),
	}, nil
}
