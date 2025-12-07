package utils

import (
	"context"
	"errors"
	"testing"
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
	ChannelsErr       error
	MessagesErr       error
}

func (f *fakeChannelsService) ListChannels(ctx context.Context, team string) ([]*channelsLib.Channel, error) {
    if f.ChannelsErr != nil {
        return nil, f.ChannelsErr
    }
    if channels, ok := f.ChannelsByTeam[team]; ok {
        return channels, nil
    }
    return []*channelsLib.Channel{}, nil
}

func (f *fakeChannelsService) ListMessages(ctx context.Context, team, channel string, opts *channelsLib.ListMessagesOptions) ([]*channelsLib.Message, error) {
    if f.MessagesErr != nil {
        return nil, f.MessagesErr
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

func getNewTestChannelMessagesClientWithError() *ChannelMessagesClient {
	fakeClient := &fakeTeamsClient{
		teamsSvc: &fakeTeamsService{
			Err: ErrListingTeamsFailed,
		},
		channelsSvc: &fakeChannelsService{
			ChannelsErr: ErrListingChannelsFailed,
			MessagesErr: ErrListingMessagesFailed,
		},
	}

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

func TestChannelMessagesClient_NoTeams(t *testing.T) {
	client := getNewTestChannelMessagesClient(false, false, false)

	_, err := client.getActiveTeams()
	if err == nil {
		t.Fatalf("Expected error when no teams are present, got nil")
	}
	if err != ErrNoTeamsFound {
		t.Fatalf("Expected ErrNoTeamsFound, got: %v", err)
	}
}

func TestChannelMessagesClient_GetActiveTeams(t *testing.T) {
	client := getNewTestChannelMessagesClient(true, false, false)

	teams, err := client.getActiveTeams()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(teams) != 1 {
		t.Fatalf("Expected 1 active team, got %d", len(teams))
	}
	if teams[0].DisplayName != "Team A" {
		t.Fatalf("Expected active team to be 'Team A', got '%s'", teams[0].DisplayName)
	}
}

func TestChannelMessagesClient_NoChannels(t *testing.T) {
	client := getNewTestChannelMessagesClient(true, false, false)
	teams, err := client.getActiveTeams()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = client.getChannels(teams)
	if err == nil {
		t.Fatalf("Expected error when no channels are present, got nil")
	}
	if err != ErrNoChannelsFound {
		t.Fatalf("Expected ErrNoChannelsFound, got: %v", err)
	}
}

func TestChannelMessagesClient_GetChannels(t *testing.T) {
	client := getNewTestChannelMessagesClient(true, true, false)
	teams, err := client.getActiveTeams()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	teamChannels, err := client.getChannels(teams)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(teamChannels) != 1 {
		t.Fatalf("Expected channels for 1 team, got %d", len(teamChannels))
	}
	channels, ok := teamChannels["Team A"]
	if !ok {
		t.Fatalf("Expected channels for 'Team A', but not found")
	}
	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels for 'Team A', got %d", len(channels))
	}
	if channels[0] != "general" || channels[1] != "party" {
		t.Fatalf("Unexpected channel names: %v", channels)
	}
}

func TestChannelMessagesClient_NoMessages(t *testing.T) {
	client := getNewTestChannelMessagesClient(true, true, false)

	messagesInfo, err := client.GetMessages()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(messagesInfo) != 0 {
		t.Fatalf("Expected 0 messages, got %d", len(messagesInfo))
	}
}

func TestChannelMessagesClient_GetMessages(t *testing.T) {
	client := getNewTestChannelMessagesClient(true, true, true)

	messagesInfo, err := client.GetMessages()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(messagesInfo) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messagesInfo))
	}
	if messagesInfo[0].TeamName != "Team A" || messagesInfo[1].TeamName != "Team A" {
		t.Fatalf("Expected messages from 'Team A', got: %v", messagesInfo)
	}
	if messagesInfo[0].ChannelName != "general" {
		t.Fatalf("Unexpected channel name for first message: %s", messagesInfo[0].ChannelName)
	}
	if messagesInfo[1].ChannelName != "party" {
		t.Fatalf("Unexpected channel name for second message: %s", messagesInfo[1].ChannelName)
	}
	if messagesInfo[0].Message.Content != "Hello, world!" && messagesInfo[1].Message.Content != "Party time!" {
		t.Fatalf("Unexpected message contents: %v", messagesInfo)
	}
}

func TestChannelMessagesClient_FetchTeamsError(t *testing.T) {
	client := getNewTestChannelMessagesClientWithError()

	_, err := client.getActiveTeams()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !errors.Is(err, ErrListingTeamsFailed) {
		t.Fatalf("Expected error to be %v, got %v", ErrListingTeamsFailed, err)
	}
}

func TestChannelMessagesClient_FetchChannelsError(t *testing.T) {
	client := getNewTestChannelMessagesClientWithError()
	teams := []*teamsLib.Team{
		{DisplayName: "Team A"},
	}

	_, err := client.getChannels(teams)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !errors.Is(err, ErrListingChannelsFailed) {
		t.Fatalf("Expected error to be %v, got %v", ErrListingChannelsFailed, err)
	}
}

func TestChannelMessagesClient_FetchMessagesError(t *testing.T) {
	client := getNewTestChannelMessagesClientWithError()

	teamChannels := TeamChannels{
		"Team A": {"general"},
	}
	
	_, err := client.getMessagesInTimeRange(teamChannels)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if !errors.Is(err, ErrListingMessagesFailed) {
		t.Fatalf("Expected error to be %v, got %v", ErrListingMessagesFailed, err)
	}
}
