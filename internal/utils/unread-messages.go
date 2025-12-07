package utils

import (
	"context"

	lib "github.com/pzsp-teams/lib"
	channelsLib "github.com/pzsp-teams/lib/channels"
	teamsLib "github.com/pzsp-teams/lib/teams"
)

type TeamChannels = map[string][]string

func filterOutArchivedTeams(teams []*teamsLib.Team) []*teamsLib.Team {
	activeTeams := teams[:0]
	for _, team := range teams {
		if !team.IsArchived {
			activeTeams = append(activeTeams, team)
		}
	}
	return activeTeams
}

func getActiveTeams(client *lib.Client) ([]*teamsLib.Team, error) {
	teams, err := client.Teams.ListMyJoined(context.TODO())
	if err != nil {
		return nil, err
	}
	teams = filterOutArchivedTeams(teams)
	return teams, nil
}

func getChannels(client *lib.Client, teams []*teamsLib.Team) (TeamChannels, error) {
	teamChannels := make(TeamChannels)
	for _, team := range teams {
		channels, err := client.Channels.ListChannels(context.TODO(), team.DisplayName)
		if err != nil {
			return nil, err
		}
		channelNames := make([]string, len(channels))
		for i, channel := range channels {
			channelNames[i] = channel.Name
		}
		teamChannels[team.DisplayName] = channelNames
	}
	return teamChannels, nil
}

func getMessagesInTimeRange(client *lib.Client, teamChannels TeamChannels, timeRange TimeRange) ([]*DisplayMessageInfo, error) {
	var messagesInfo []*DisplayMessageInfo
	top := int32(100)
	opts := &channelsLib.ListMessagesOptions{
		Top:           &top,
		ExpandReplies: true,
	}

	for team, channels := range teamChannels {
		for _, channel := range channels {
			messages, err := client.Channels.ListMessages(context.TODO(), team, channel, opts)
			if err != nil {
				return nil, err
			}

			for _, message := range messages {
				if message.CreatedDateTime.After(timeRange.Start) && message.CreatedDateTime.Before(timeRange.End) {
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
