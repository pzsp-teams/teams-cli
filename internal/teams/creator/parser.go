package creator

import (
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// ParseTeamsData parses team data using the provided decode function.
func ParseTeamsData(r io.Reader, decodeFunc file_readers.DecodeFunc) (map[string]TeamData, error) {
	teamsData := make(map[string]TeamData, 0)
	err := decodeFunc(r, &teamsData)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}
	initializers.Logger.Info("Teams data parsed", "team_count", len(teamsData))
	return teamsData, nil
}

// ParseTeamsDataByExtension parses team data auto-detecting format by file extension.
// Supported extensions: yaml, yml, json, toml, csv.
func ParseTeamsDataByExtension(r io.Reader, extension string) (map[string]TeamData, error) {
	if extension == "csv" {
		return parseTeamsDataFromCSV(r)
	}

	decoder, err := file_readers.GetDecoderByExtension(extension)
	if err != nil {
		return nil, err
	}

	return ParseTeamsData(r, decoder)
}

func parseTeamsDataFromCSV(r io.Reader) (map[string]TeamData, error) {
	var rows []map[string]string
	err := file_readers.DecodeCSV(r, &rows)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}

	teamsData := transformCSVRowsToTeamData(rows)
	initializers.Logger.Info("Teams data parsed from CSV", "team_count", len(teamsData))
	return teamsData, nil
}

func transformCSVRowsToTeamData(rows []map[string]string) map[string]TeamData {
	grouped := file_readers.GroupBy(rows, func(row map[string]string) string {
		return row["team_ref"]
	})

	result := make(map[string]TeamData, len(grouped))
	for teamRef, teamRows := range grouped {
		teamData := TeamData{
			"members":     []string{},
			"owners":      []string{},
			"description": "",
		}

		for i, row := range teamRows {
			if i == 0 {
				teamData["description"] = row["description"]
			} else if row["description"] != "" && row["description"] != teamData["description"].(string) {
				initializers.Logger.Warn("Inconsistent description for team", "team", teamRef, "first", teamData["description"], "row", row["description"])
			}

			role := row["role"]
			userRef := row["user_ref"]
			switch role {
			case "member":
				members := teamData["members"].([]string)
				teamData["members"] = append(members, userRef)
			case "owner":
				owners := teamData["owners"].([]string)
				teamData["owners"] = append(owners, userRef)
			}
		}
		result[teamRef] = teamData
	}

	return result
}
