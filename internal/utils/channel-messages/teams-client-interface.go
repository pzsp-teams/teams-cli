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

// Wrap real client to implement TeamsClient interface
type realTeamsClient struct {
	client *lib.Client
}

func (r *realTeamsClient) Teams() TeamsService {
	return &realTeamsService{r.client.Teams}
}

func (r *realTeamsClient) Channels() ChannelsService {
	return &realChannelsService{r.client.Channels}
}

type realTeamsService struct {
	svc *teamsLib.Service
}

func (r *realTeamsService) ListMyJoined(ctx context.Context) ([]*teamsLib.Team, error) {
	return r.svc.ListMyJoined(ctx)
}

type realChannelsService struct {
	svc *channelsLib.Service
}

func (r *realChannelsService) ListChannels(ctx context.Context, team string) ([]*channelsLib.Channel, error) {
	return r.svc.ListChannels(ctx, team)
}

func (r *realChannelsService) ListMessages(ctx context.Context, team, channel string, opts *channelsLib.ListMessagesOptions) ([]*channelsLib.Message, error) {
	return r.svc.ListMessages(ctx, team, channel, opts)
}

func wrapRealClient(client *lib.Client) TeamsClient {
	return &realTeamsClient{client}
}
