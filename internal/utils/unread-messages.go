package utils

import (
	"context"

	lib "github.com/pzsp-teams/lib"
	teamsLib "github.com/pzsp-teams/lib/teams"
	channelsLib "github.com/pzsp-teams/lib/channels"
)

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

func getChannels(client *lib.Client, teams []*teamsLib.Team) (map[*teamsLib.Team][]*channelsLib.Channel, error) {
	teamChannels := make(map[*teamsLib.Team][]*channelsLib.Channel);
	for _, team := range teams {
		channels, err := client.Channels.ListChannels(context.TODO(), team.DisplayName)
		if err != nil {
			return nil, err
		}
		teamChannels[team] = channels
	}
	return teamChannels, nil
}

