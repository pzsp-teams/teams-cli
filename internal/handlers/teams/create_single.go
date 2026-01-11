package teams

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"

	"github.com/pzsp-teams/cli/internal/client"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/core/creator"
	"github.com/pzsp-teams/cli/internal/initializers"
	tcreator "github.com/pzsp-teams/cli/internal/teams/creator"
	"github.com/pzsp-teams/cli/internal/teams/creator/single"
)

// CreateSingleTeam handles creating a single team.
func CreateSingleTeam(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	teamName, _ := flags["team-name"].(string)
	description, _ := flags["description"].(string)
	visibility, _ := flags["visibility"].(string)
	file, _ := flags["file"].(string)
	includeMe, _ := flags["include-me"].(bool)
	dryRun, _ := flags["dry-run"].(bool)

	// Validation
	validate := validator.New()
	if err := validate.Var(teamName, "required"); err != nil {
		return nil, fmt.Errorf("team-name is required")
	}
	if err := validate.Var(file, "required"); err != nil {
		return nil, fmt.Errorf("file is required")
	}

	logger := initializers.Logger

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file, err)
	}

	ext := internalcommon.GetFileExtension(file)
	singleTeamData, err := single.ParseSingleTeamDataByExtension(f, ext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse team data from %s: %w", file, err)
	}

	_ = f.Close()

	teamData := tcreator.TeamData{
		Description: description,
		Owners:      singleTeamData.Owners,
		Members:     singleTeamData.Members,
		Visibility:  visibility,
		IncludeMe:   includeMe,
	}

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	request := map[string]tcreator.TeamData{
		teamName: teamData,
	}

	results := c.Teams.Create(ctx, request, dryRun)

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
		return nil, fmt.Errorf("one or more teams failed to create")
	}

	return results, nil
}
