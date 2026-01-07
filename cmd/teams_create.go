package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pzsp-teams/cli/internal/initializers"
	teamcreator "github.com/pzsp-teams/cli/internal/teams/creator"
	"github.com/spf13/cobra"
)

var teamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create teams from a data file",
	Long: `Create multiple Teams from a data file (YAML/JSON/CSV).

The data file should contain team definitions with display names, descriptions, owners, and members.

Examples:
  # Create teams from YAML file
  cli team create --data teams.yaml

  # Create teams from JSON file
  cli team create --data teams.json

  # Dry run to preview
  cli team create --data teams.yaml --dry-run
`,
	RunE: runTeamCreate,
}

var (
	createTeamsData   string
	createTeamsDryRun bool
)

func runTeamCreate(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := cmd.Context()

	dataFile, err := os.Open(createTeamsData)
	if err != nil {
		log.Error("Failed to open data file", "file", createTeamsData, "error", err)
		return fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(createTeamsData), ".")

	log.Debug("Parsing teams data", "file", createTeamsData)
	teamData, err := teamcreator.ParseTeamsDataByExtension(dataFile, extension)
	if err != nil {
		log.Error("Failed to parse teams data", "error", err)
		return fmt.Errorf("failed to parse teams data: %w", err)
	}

	_ = dataFile.Close()

	log.Info("Parsed teams data", "teams", len(teamData))

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Creating teams", "count", len(teamData), "dryRun", createTeamsDryRun)
	results := teamsClient.Teams.Create(ctx, teamData, createTeamsDryRun)

	printTeamCreationResults(results, createTeamsDryRun)

	return nil
}

func printTeamCreationResults(results []teamcreator.TeamCreateResult, dryRun bool) {
	successCount := 0
	for i := range results {
		res := &results[i]
		if res.Error != nil {
			fmt.Printf("Failed - team: %s, error: %v\n", res.TeamName, res.Error)
		} else {
			successCount++
			if dryRun {
				fmt.Printf("[Dry Run] Would create - team: %s\n", res.TeamName)
			} else {
				fmt.Printf("Created - team: %s (ID: %s)\n", res.TeamName, res.TeamID)
			}
			if res.Description != "" {
				fmt.Printf("  Description: %s\n", res.Description)
			}
			if len(res.OwnerRefs) > 0 {
				fmt.Printf("  Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
			}
			if len(res.MemberRefs) > 0 {
				fmt.Printf("  Members: %s\n", strings.Join(res.MemberRefs, ", "))
			}
			fmt.Printf("  Visibility: %s\n", res.Visibility)
		}
	}

	if dryRun {
		fmt.Printf("\nDry run completed - successful: %d, total: %d\n", successCount, len(results))
	} else {
		fmt.Printf("\nTeam creation completed - successful: %d, total: %d\n", successCount, len(results))
	}
}
