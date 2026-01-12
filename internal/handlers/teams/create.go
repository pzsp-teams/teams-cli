package teams

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pzsp-teams/teams-cli/internal/client"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
	teamcreator "github.com/pzsp-teams/teams-cli/internal/teams/creator"
)

// CreateTeams handles creating teams from a file.
func CreateTeams(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	data, _ := flags["data"].(string)
	dryRun, _ := flags["dry-run"].(bool)

	log := initializers.Logger

	dataFile, err := os.Open(data)
	if err != nil {
		log.Error("Failed to open data file", "file", data, "error", err)
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	extension := strings.TrimPrefix(filepath.Ext(data), ".")

	log.Debug("Parsing teams data", "file", data)
	teamData, err := teamcreator.ParseTeamsDataByExtension(dataFile, extension)
	if err != nil {
		log.Error("Failed to parse teams data", "error", err)
		return nil, fmt.Errorf("failed to parse teams data: %w", err)
	}

	_ = dataFile.Close()

	log.Info("Parsed teams data", "teams", len(teamData))

	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	log.Info("Creating teams", "count", len(teamData), "dryRun", dryRun)
	results := c.Teams.Create(ctx, teamData, dryRun)

	printTeamCreationResults(w, results, dryRun)

	return results, nil
}

func printTeamCreationResults(w io.Writer, results []teamcreator.TeamCreateResult, dryRun bool) {
	successCount := 0
	for i := range results {
		res := &results[i]
		if res.Error != nil {
			_, _ = fmt.Fprintf(w, "Failed - team: %s, error: %v\n", res.TeamName, res.Error)
		} else {
			successCount++
			if dryRun {
				_, _ = fmt.Fprintf(w, "[Dry Run] Would create - team: %s\n", res.TeamName)
			} else {
				_, _ = fmt.Fprintf(w, "Created - team: %s (ID: %s)\n", res.TeamName, res.TeamID)
			}
			if res.Description != "" {
				_, _ = fmt.Fprintf(w, "  Description: %s\n", res.Description)
			}
			if len(res.OwnerRefs) > 0 {
				_, _ = fmt.Fprintf(w, "  Owners: %s\n", strings.Join(res.OwnerRefs, ", "))
			}
			if len(res.MemberRefs) > 0 {
				_, _ = fmt.Fprintf(w, "  Members: %s\n", strings.Join(res.MemberRefs, ", "))
			}
			_, _ = fmt.Fprintf(w, "  Visibility: %s\n", res.Visibility)
		}
	}

	if dryRun {
		_, _ = fmt.Fprintf(w, "\nDry run completed - successful: %d, total: %d\n", successCount, len(results))
	} else {
		_, _ = fmt.Fprintf(w, "\nTeam creation completed - successful: %d, total: %d\n", successCount, len(results))
	}
}
