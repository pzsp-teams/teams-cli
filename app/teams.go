package app

import (
	teamsHandlers "github.com/pzsp-teams/cli/internal/handlers/teams"
)

var (
	teamsLong = `Commands for interacting with Microsoft Teams teams`

	teamsListLong = `Retrieve and display a list of all Microsoft Teams that the authenticated user is a member of.`

	teamsGetLong = `Retrieve and display details of a specific Microsoft Teams team by its ID or display name.`

	teamsCreateLong = `Create multiple Teams from a data file (YAML/JSON/CSV).

The data file should contain team definitions with display names, descriptions, owners, and members.

Examples:
  # Create teams from YAML file
  teams-cli teams create --data teams.yaml

  # Create teams from JSON file
  teams-cli teams create --data teams.json

  # Dry run to preview
  teams-cli teams create --data teams.yaml --dry-run`

	teamsCreateSingleLong = `This command creates a single team based on provided flags and a simplified data file.`

	teamsArchiveLong = `Archive a Microsoft Teams team by its ID or display name.`

	teamsUnarchiveLong = `Unarchive a Microsoft Teams team by its ID or display name.`

	teamsDeleteLong = `Delete a Microsoft Teams team by its ID or display name.`
)

func init() {
	teamsCmd := CommandDef{
		Use:   "teams",
		Short: "Manage teams",
		Long:  teamsLong,
		SubCommands: []CommandDef{
			{
				Use:     "list",
				Short:   "List all teams",
				Long:    teamsListLong,
				Handler: teamsHandlers.ListTeams,
			},
			{
				Use:   "get",
				Short: "Get team details",
				Long:  teamsGetLong,
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.GetTeam,
			},
			{
				Use:   "create",
				Short: "Create teams from file",
				Long:  teamsCreateLong,
				Flags: []FlagDef{
					{Name: "data", Usage: "Path to teams data file (YAML/JSON/CSV)", Type: InputFile, Required: true},
					{Name: "dry-run", Usage: "Preview without creating teams", Type: InputBool},
				},
				Handler: teamsHandlers.CreateTeams,
			},
			{
				Use:   "create-single",
				Short: "Create single team",
				Long:  teamsCreateSingleLong,
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
				Long:  teamsArchiveLong,
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
					{Name: "spo-read-only", Usage: "Set SharePoint to read-only", Type: InputBool},
				},
				Handler: teamsHandlers.ArchiveTeam,
			},
			{
				Use:   "unarchive",
				Short: "Unarchive a team",
				Long:  teamsUnarchiveLong,
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.UnarchiveTeam,
			},
			{
				Use:   "delete",
				Short: "Delete a team",
				Long:  teamsDeleteLong,
				Flags: []FlagDef{
					{Name: "team", Usage: "ID or display name of the team", Type: InputString, Required: true},
				},
				Handler: teamsHandlers.DeleteTeam,
			},
		},
	}

	Registry = append(Registry, teamsCmd)
}
