package utils

import (
	"context"
	"time"

	channelsLib "github.com/pzsp-teams/lib/channels"
	teamsLib "github.com/pzsp-teams/lib/teams"
)

type fakeTeamsService struct {
	TeamsToReturn []*teamsLib.Team
	Err           error
}

func (f *fakeTeamsService) ListMyJoined(ctx context.Context) ([]*teamsLib.Team, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	return f.TeamsToReturn, nil
}

type fakeChannelsService struct {
	ChannelsByTeam    map[string][]*channelsLib.Channel
	MessagesByChannel map[string][]*channelsLib.Message
	Err               error
}

func (f *fakeChannelsService) ListChannels(ctx context.Context, team string) ([]*channelsLib.Channel, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	if channels, ok := f.ChannelsByTeam[team]; ok {
		return channels, nil
	}
	return []*channelsLib.Channel{}, nil
}

func (f *fakeChannelsService) ListMessages(ctx context.Context, team, channel string, opts *channelsLib.ListMessagesOptions) ([]*channelsLib.Message, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	if messages, ok := f.MessagesByChannel[channel]; ok {
		return messages, nil
	}
	return []*channelsLib.Message{}, nil
}

type fakeTeamsClient struct {
	teamsSvc    TeamsService
	channelsSvc ChannelsService
}

func (f *fakeTeamsClient) Teams() TeamsService {
	return f.teamsSvc
}

func (f *fakeTeamsClient) Channels() ChannelsService {
	return f.channelsSvc
}

func newFakeTeamsClientWithInterface(hasTeams, hasChannels, hasMessages bool) *fakeTeamsClient {
	teamsSvc := &fakeTeamsService{}
	channelsSvc := &fakeChannelsService{
		ChannelsByTeam:    map[string][]*channelsLib.Channel{},
		MessagesByChannel: map[string][]*channelsLib.Message{},
	}

	if hasTeams {
		teamsSvc.TeamsToReturn = []*teamsLib.Team{
			{DisplayName: "Team A", IsArchived: false},
			{DisplayName: "Team B", IsArchived: true},
		}
	}

	if hasChannels {
		channelsSvc.ChannelsByTeam = map[string][]*channelsLib.Channel{
			"Team A": {
				{Name: "general"},
				{Name: "party"},
			},
			"Team B": {
				{Name: "general B"},
			},
		}
	}

	if hasMessages {
		channelsSvc.MessagesByChannel = map[string][]*channelsLib.Message{
			"general": {
				{Content: "Hello, world!"},
			},
			"party": {
				{Content: "Party time!"},
			},
			"general B": {
				{Content: "This is Team B."},
			},
		}
	}

	return &fakeTeamsClient{
		teamsSvc:    teamsSvc,
		channelsSvc: channelsSvc,
	}
}

func getNewTestChannelMessagesClient(hasTeams, hasChannels, hasMessages bool) *ChannelMessagesClient {
	fakeClient := newFakeTeamsClientWithInterface(hasTeams, hasChannels, hasMessages)

	timeRange := TimeRange{
		Start: time.Unix(0, 0),
		End:   time.Now().Add(24 * time.Hour),
	}

	return &ChannelMessagesClient{
		teamsClient: fakeClient,
		TimeRange:   timeRange,
	}
}

// Actual tests


