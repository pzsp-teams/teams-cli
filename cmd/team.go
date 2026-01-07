package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corecreator "github.com/pzsp-teams/cli/internal/core/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
	teamcreator "github.com/pzsp-teams/cli/internal/teams/creator"
	"github.com/pzsp-teams/lib/models"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Simple team operations",
	Long:  `Commands for simple team operations on Microsoft Teams`,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams the user is a member of",
	Long:  `Retrieve and display a list of all Microsoft Teams that the authenticated user is a member of.`,
	RunE:  runTeamList,
}

var (
	tRef        string
	spoReadOnly bool
)

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamGetCmd)
	teamGetCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to get details for")
	if err := teamGetCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamCreateCmd)
	teamCreateCmd.Flags().StringVar(&createTeamsData, "data", "", "Path to teams data file (YAML/JSON/CSV)")
	teamCreateCmd.Flags().BoolVar(&createTeamsDryRun, "dry-run", false, "Preview without creating teams")
	if err := teamCreateCmd.MarkFlagRequired("data"); err != nil {
		panic(fmt.Sprintf("failed to mark data flag as required: %v", err))
	}

	teamCmd.AddCommand(teamArchiveCmd)
	teamArchiveCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to archive")
	teamArchiveCmd.Flags().BoolVar(&spoReadOnly, "spo-read-only", false, "Set SharePoint Online site to read-only mode when archiving the team")
	if err := teamArchiveCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamUnarchiveCmd)
	teamUnarchiveCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to unarchive")
	if err := teamUnarchiveCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamDeleteCmd)
	teamDeleteCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to delete")
	if err := teamDeleteCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runTeamList(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	ts, err := c.Teams.ListMyJoined(cmd.Context())
	if err != nil {
		return err
	}
	if len(ts) == 0 {
		cmd.Println("No teams found.")
		return nil
	}
	fmt.Println("Teams:")
	for _, t := range ts {
		state := ""
		if t.IsArchived {
			state = " (Archived)"
		}
		fmt.Printf("- %s%s (ID: %s)\n", t.DisplayName, state, t.ID)
	}
	return nil
}

var teamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a specific team",
	Long:  `Retrieve and display details of a specific Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamGet,
}

func runTeamGet(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	t, err := c.Teams.Get(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	printTeamDetails(t)
	return nil
}

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
				switch res.Status {
				case corecreator.StatusWouldCreate:
					fmt.Printf("[Dry Run] Would create - team: %s\n", res.TeamName)
					if res.Description != "" {
						fmt.Printf("  Description: %s\n", res.Description)
					}
					if len(res.OwnerRefs) > 0 {
						fmt.Printf("  Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
					}
					if len(res.MemberRefs) > 0 {
						fmt.Printf("  Members: %s\n", strings.Join(res.MemberRefs, ", "))
					}
				case corecreator.StatusAlreadyExists:
					fmt.Printf("[Dry Run] Already exists - team: %s\n", res.TeamName)
				default:
					fmt.Printf("[Dry Run] Processed - team: %s\n", res.TeamName)
				}
			} else {
				switch res.Status {
				case corecreator.StatusCreated:
					fmt.Printf("Created - team: %s (ID: %s)\n", res.TeamName, res.TeamID)
					if res.Description != "" {
						fmt.Printf("  Description: %s\n", res.Description)
					}
					if len(res.OwnerRefs) > 0 {
						fmt.Printf("  Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
					}
					if len(res.MemberRefs) > 0 {
						fmt.Printf("  Members: %s\n", strings.Join(res.MemberRefs, ", "))
					}
				case corecreator.StatusAlreadyExists:
					fmt.Printf("Already exists - team: %s\n", res.TeamName)
				default:
					fmt.Printf("Processed - team: %s\n", res.TeamName)
				}
			}
		}
	}

	if dryRun {
		fmt.Printf("\nDry run completed - successful: %d, total: %d\n", successCount, len(results))
	} else {
		fmt.Printf("\nTeam creation completed - successful: %d, total: %d\n", successCount, len(results))
	}
}

func printTeamDetails(t *models.Team) {
	fmt.Printf("Team Details:\n")
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Display Name: %s\n", t.DisplayName)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Is Archived: %t\n", t.IsArchived)
	if t.Visibility != nil {
		fmt.Printf("Visibility: %s\n", *t.Visibility)
	}
}

var teamArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a team",
	Long:  `Archive a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamArchive,
}

func runTeamArchive(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Archive(cmd.Context(), tRef, spoReadOnly)
	if err != nil {
		return err
	}
	fmt.Println("Team archive initiated. The team will be archived shortly.")
	return nil
}

var teamUnarchiveCmd = &cobra.Command{
	Use:   "unarchive",
	Short: "Unarchive a team",
	Long:  `Unarchive a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamUnarchive,
}

func runTeamUnarchive(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Unarchive(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	fmt.Println("Team unarchive initiated. The team will be unarchived shortly.")
	return nil
}

var teamDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a team",
	Long:  `Delete a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamDelete,
}

func runTeamDelete(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Delete(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	fmt.Println("Team deletion initiated. The team will be deleted shortly.")
	return nil
}
