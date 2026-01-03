package client

import (
	"context"

	"github.com/pzsp-teams/cli/internal/channels"
	"github.com/pzsp-teams/cli/internal/chats"
	"github.com/pzsp-teams/lib"
	lib_config "github.com/pzsp-teams/lib/config"
)

// TeamsClient aggregates resource-level clients
type TeamsClient struct {
	Client   *lib.Client
	Channels *channels.Client
	Chats    *chats.Client
}

// NewTeamsClient creates a new TeamsClient by constructing the underlying library client
func NewTeamsClient(ctx context.Context, authConfig *lib_config.AuthConfig, senderConfig *lib_config.SenderConfig, cacheConfig *lib_config.CacheConfig) (*TeamsClient, error) {
	libClient, err := lib.NewClient(ctx, authConfig, senderConfig, cacheConfig)
	if err != nil {
		return nil, err
	}

	return &TeamsClient{
		Client:   libClient,
		Channels: channels.NewClient(libClient.Channels, libClient.Teams),
		Chats:    chats.NewClient(libClient.Chats),
	}, nil
}
