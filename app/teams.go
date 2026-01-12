package app

import (
	teamsHandlers "github.com/pzsp-teams/cli/internal/handlers/teams"
)

func init() {
	teamsCmd := CommandDef{
		Use:   "teams",
		Short: "Manage teams",
		SubCommands: []CommandDef{
			{
				Use:     "list",
				Short:   "List all teams",
				Handler: teamsHandlers.ListTeams,
			},
			{
				Use:   "get",
				Short: "Get team details",
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.GetTeam,
			},
			{
				Use:   "create",
				Short: "Create teams from file",
				Flags: []FlagDef{
					{Name: "data", Usage: "Path to teams data file (YAML/JSON/CSV)", Type: InputFile, Required: true},
					{Name: "dry-run", Usage: "Preview without creating teams", Type: InputBool},
				},
				Handler: teamsHandlers.CreateTeams,
			},
			{
				Use:   "create-single",
				Short: "Create single team",
				Flags: []FlagDef{
					{Name: "team-name", Usage: "Name of the team to create", Type: InputString, Required: true},
					{Name: "description", Usage: "Description of the team", Type: InputString},
					{
						Name:       "visibility",
						Usage:      "Visibility (private/public)",
						Type:       InputChoice,
						DefaultVal: "private",
						Options:    []string{"private", "public"},
					},
					{Name: "file", Shorthand: "f", Usage: "Path to members data file", Type: InputFile, Required: true},
					{Name: "include-me", Shorthand: "i", Usage: "Include current user", Type: InputBool},
					{Name: "dry-run", Usage: "Preview only", Type: InputBool},
				},
				Handler: teamsHandlers.CreateSingleTeam,
			},
			{
				Use:   "archive",
				Short: "Archive a team",
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
					{Name: "spo-read-only", Usage: "Set SharePoint to read-only", Type: InputBool},
				},
				Handler: teamsHandlers.ArchiveTeam,
			},
			{
				Use:   "unarchive",
				Short: "Unarchive a team",
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.UnarchiveTeam,
			},
			{
				Use:   "delete",
				Short: "Delete a team",
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.DeleteTeam,
			},
		},
	}

	Registry = append(Registry, teamsCmd)
}
