package utils

import (
	"context"

	lib "github.com/pzsp-teams/lib"
	channelsLib "github.com/pzsp-teams/lib/channels"
	teamsLib "github.com/pzsp-teams/lib/teams"
)

type TeamsClient interface {
	Teams() TeamsService
	Channels() ChannelsService
}

type TeamsService interface {
	ListMyJoined(ctx context.Context) ([]*teamsLib.Team, error)
}

type ChannelsService interface {
	ListChannels(ctx context.Context, team string) ([]*channelsLib.Channel, error)
	ListMessages(ctx context.Context, team, channel string, opts *channelsLib.ListMessagesOptions) ([]*channelsLib.Message, error)
}

type clientAdapter struct {
	client *lib.Client
}

func (c *clientAdapter) Teams() TeamsService {
	return c.client.Teams
}

func (c *clientAdapter) Channels() ChannelsService {
	return c.client.Channels
}

func wrapRealClient(client *lib.Client) TeamsClient {
	return &clientAdapter{client}
}
