package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/cli/internal/core/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
	tcreator "github.com/pzsp-teams/cli/internal/teams/creator"
	"github.com/pzsp-teams/cli/internal/teams/creator/single"
)

type createSingleTeamCommand struct {
	teamName    string
	description string
	visibility  string
	file        string
	includeMe   bool
	dryRun      bool
}

func newCreateSingleTeamCommand() *cobra.Command {
	c := &createSingleTeamCommand{}
	cmd := &cobra.Command{
		Use:   "create-single",
		Short: "Create a single team",
		Long:  "This command creates a single team based on provided flags and a simplified data file.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			validate := validator.New()
			if err := validate.Var(c.teamName, "required"); err != nil {
				return fmt.Errorf("team-name is required")
			}
			if err := validate.Var(c.file, "required"); err != nil {
				return fmt.Errorf("file is required")
			}
			return nil
		},
		RunE: c.run,
	}

	cmd.Flags().StringVarP(&c.teamName, "team-name", "", "", "Name of the team to create")
	cmd.Flags().StringVarP(&c.description, "description", "", "", "Description of the team")
	cmd.Flags().StringVarP(&c.visibility, "visibility", "", "private", "Visibility of the team (e.g., 'private', 'public')")
	cmd.Flags().StringVarP(&c.file, "file", "f", "", "Path to the simplified data file (CSV, YAML, JSON, TOML)")
	cmd.Flags().BoolVarP(&c.includeMe, "include-me", "i", false, "Include current user as a member of the team")
	cmd.Flags().BoolVarP(&c.dryRun, "dry-run", "", false, "Perform a dry run without actually creating teams")

	return cmd
}

func (c *createSingleTeamCommand) run(cmd *cobra.Command, args []string) error {
	logger := initializers.Logger

	f, err := os.Open(c.file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", c.file, err)
	}

	ext := getFileExtension(c.file)
	singleTeamData, err := single.ParseSingleTeamDataByExtension(f, ext)
	if err != nil {
		return fmt.Errorf("failed to parse team data from %s: %w", c.file, err)
	}

	_ = f.Close()

	teamData := tcreator.TeamData{
		Description: c.description,
		Owners:      singleTeamData.Owners,
		Members:     singleTeamData.Members,
		Visibility:  c.visibility,
		IncludeMe:   c.includeMe,
	}

	teamsClient, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	request := map[string]tcreator.TeamData{
		c.teamName: teamData,
	}

	results := teamsClient.Teams.Create(cmd.Context(), request, c.dryRun)

	var hasErrors bool
	for i := range results {
		result := &results[i]
		if result.Status == creator.StatusFailed {
			hasErrors = true
			logger.Error("Failed to create single team", "team", result.TeamName, "error", result.Error)
		} else {
			logger.Info("Single team creation result", "team", result.TeamName, "status", result.Status, "team_id", result.TeamID)
		}
	}

	if hasErrors {
		return fmt.Errorf("one or more teams failed to create")
	}

	return nil
}
