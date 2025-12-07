package utils

import (
	"context"

	lib "github.com/pzsp-teams/lib"
	channelsLib "github.com/pzsp-teams/lib/channels"
	teamsLib "github.com/pzsp-teams/lib/teams"
)

type teamsClient interface {
	teams() teamsService
	channels() channelsService
}

type teamsService interface {
	ListMyJoined(ctx context.Context) ([]*teamsLib.Team, error)
}

type channelsService interface {
	ListChannels(ctx context.Context, team string) ([]*channelsLib.Channel, error)
	ListMessages(ctx context.Context, team, channel string, opts *channelsLib.ListMessagesOptions) ([]*channelsLib.Message, error)
}

type clientAdapter struct {
	client *lib.Client
}

func (c *clientAdapter) teams() teamsService {
	return c.client.Teams
}

func (c *clientAdapter) channels() channelsService {
	return c.client.Channels
}

func wrapRealClient(client *lib.Client) teamsClient {
	return &clientAdapter{client}
}
