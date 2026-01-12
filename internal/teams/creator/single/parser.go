package single

import (
	"fmt"
	"io"

	"github.com/pzsp-teams/teams-cli/internal/file_readers"
	"github.com/pzsp-teams/teams-cli/internal/initializers"
)

// ParseSingleTeamDataByExtension parses team data for a single team
// auto-detecting format by file extension.
// Supported extensions: yaml, yml, json, toml, csv.
func ParseSingleTeamDataByExtension(r io.Reader, extension string) (TeamData, error) {
	if extension == "csv" {
		return parseSingleTeamDataFromCSV(r)
	}

	decoder, err := file_readers.GetDecoderByExtension(extension)
	if err != nil {
		return TeamData{}, err
	}

	var teamData TeamData
	err = decoder(r, &teamData)
	if err != nil {
		initializers.Logger.Error("failed to parse single team data", "error", err)
		return TeamData{}, fmt.Errorf("failed to parse single team data: %w", err)
	}
	initializers.Logger.Info("Single team data parsed")
	return teamData, nil
}

func parseSingleTeamDataFromCSV(r io.Reader) (TeamData, error) {
	var rows []map[string]string
	err := file_readers.DecodeCSV(r, &rows)
	if err != nil {
		initializers.Logger.Error("failed to parse single team CSV data", "error", err)
		return TeamData{}, fmt.Errorf("failed to parse single team CSV data: %w", err)
	}

	teamData := transformCSVRowsToSingleTeamData(rows)
	initializers.Logger.Info("Single team data parsed from CSV")
	return teamData, nil
}

func transformCSVRowsToSingleTeamData(rows []map[string]string) TeamData {
	teamData := TeamData{
		Members: []string{},
		Owners:  []string{},
	}

	for _, row := range rows {
		role := row["role"]
		userRef := row["user_ref"]
		switch role {
		case "member":
			teamData.Members = append(teamData.Members, userRef)
		case "owner":
			teamData.Owners = append(teamData.Owners, userRef)
		}
	}
	return teamData
}
