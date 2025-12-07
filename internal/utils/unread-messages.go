package utils

import (
	"context"

	lib "github.com/pzsp-teams/lib"
	teamsLib "github.com/pzsp-teams/lib/teams"
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

// func getChannels(client )