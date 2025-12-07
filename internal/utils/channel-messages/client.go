package utils

import (
	"context"
	"fmt"

	lib "github.com/pzsp-teams/lib"
	channelsLib "github.com/pzsp-teams/lib/channels"
	teamsLib "github.com/pzsp-teams/lib/teams"
)

type teamChannels = map[string][]string

// ChannelMessagesClient will be used later
type ChannelMessagesClient struct {
	teamsClient teamsClient
	TimeRange   TimeRange
}

// NewChannelMessagesClient will be used later
func NewChannelMessagesClient(client *lib.Client, timeRange TimeRange) *ChannelMessagesClient {
	return &ChannelMessagesClient{
		teamsClient: wrapRealClient(client),
		TimeRange:   timeRange,
	}
}

func (c *ChannelMessagesClient) filterOutArchivedTeams(teams []*teamsLib.Team) []*teamsLib.Team {
	activeTeams := teams[:0]
	for _, team := range teams {
		if !team.IsArchived {
			activeTeams = append(activeTeams, team)
		}
	}
	return activeTeams
}

func (c *ChannelMessagesClient) getActiveTeams() ([]*teamsLib.Team, error) {
	teams, err := c.teamsClient.teams().ListMyJoined(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListingTeamsFailed, err)
	}
	teams = c.filterOutArchivedTeams(teams)
	if len(teams) == 0 {
		return nil, ErrNoTeamsFound
	}
	return teams, nil
}

func (c *ChannelMessagesClient) getChannels(teams []*teamsLib.Team) (teamChannels, error) {
	teamChannels := make(teamChannels)
	for _, team := range teams {
		channels, err := c.teamsClient.channels().ListChannels(context.TODO(), team.DisplayName)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrListingChannelsFailed, team.DisplayName, err)
		}
		channelNames := make([]string, len(channels))
		for i, channel := range channels {
			channelNames[i] = channel.Name
		}
		if len(channelNames) > 0 {
			teamChannels[team.DisplayName] = channelNames
		}
	}
	if len(teamChannels) == 0 {
		return nil, ErrNoChannelsFound
	}
	return teamChannels, nil
}

func (c *ChannelMessagesClient) getMessagesInTimeRange(teamChannels teamChannels) ([]*DisplayMessageInfo, error) {
	var messagesInfo []*DisplayMessageInfo
	top := int32(100)
	opts := &channelsLib.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}

	for team, channels := range teamChannels {
		for _, channel := range channels {
			messages, err := c.teamsClient.channels().ListMessages(context.TODO(), team, channel, opts)
			if err != nil {
				return nil, fmt.Errorf("%w: team=%s channel=%s: %v",
					ErrListingMessagesFailed, team, channel, err)
			}

			for _, message := range messages {
				if message.CreatedDateTime.After(c.TimeRange.Start) && message.CreatedDateTime.Before(c.TimeRange.End) {
					messagesInfo = append(messagesInfo, &DisplayMessageInfo{
						TeamName:    team,
						ChannelName: channel,
						Message:     message,
					})
				}
			}
		}
	}

	return messagesInfo, nil
}

// GetMessages will be used later
func (c *ChannelMessagesClient) GetMessages() ([]*DisplayMessageInfo, error) {
	teams, err := c.getActiveTeams()
	if err != nil {
		return nil, err
	}
	// TODO: Apply Teams filters here
	teamChannels, err := c.getChannels(teams)
	if err != nil {
		return nil, err
	}
	//TODO: Apply Channels filters here
	messagesInfo, err := c.getMessagesInTimeRange(teamChannels)
	if err != nil {
		return nil, err
	}
	return messagesInfo, nil
}
