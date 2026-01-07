package creator

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/stretchr/testify/require"
)

func TestParseTeamsData_WhenDecodeFails_ReturnsWrappedError(t *testing.T) {
	origErr := errors.New("boom")

	decodeFn := file_readers.DecodeFunc(func(_ io.Reader, _ any) error {
		return origErr
	})

	got, err := ParseTeamsData(bytes.NewBufferString("x"), decodeFn)
	require.Nil(t, got)
	require.Error(t, err)

	require.ErrorIs(t, err, errDataParseFailed)
	require.ErrorIs(t, err, origErr)
}

func TestParseTeamsData_WhenDecodeOK_ReturnsData(t *testing.T) {
	expected := map[string]TeamData{
		"team1": {
			Description: "Team 1 description",
			Members:     []string{"u1"},
			Owners:      []string{"u2"},
		},
		"team2": {
			Description: "Team 2 description",
			Members:     []string{},
			Owners:      []string{"u3", "u4"},
		},
	}

	decodeFn := file_readers.DecodeFunc(func(_ io.Reader, v any) error {
		ptr, ok := v.(*map[string]TeamData)
		require.True(t, ok, "decodeFn should receive *map[string]TeamData")
		*ptr = expected
		return nil
	})

	got, err := ParseTeamsData(bytes.NewBufferString("x"), decodeFn)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestParseTeamsDataFromCSV_Success(t *testing.T) {
	csv := "team_ref,role,user_ref\ntest-team1,owner,kmarszalek@teamspzsp.onmicrosoft.com\ntest-team1,member,ddsouza@teamspzsp.onmicrosoft.com\ntest-team2,owner,msuski@teamspzsp.onmicrosoft.com\ntest-team2,owner,kmarszalek@teamspzsp.onmicrosoft.com\n"

	got, err := parseTeamsDataFromCSV(bytes.NewBufferString(csv))
	require.NoError(t, err)

	require.Len(t, got, 2)
	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{"ddsouza@teamspzsp.onmicrosoft.com"},
		Owners:      []string{"kmarszalek@teamspzsp.onmicrosoft.com"},
		Visibility:  "private",
	}, got["test-team1"])

	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{},
		Owners:      []string{"msuski@teamspzsp.onmicrosoft.com", "kmarszalek@teamspzsp.onmicrosoft.com"},
		Visibility:  "private",
	}, got["test-team2"])
}

func TestParseTeamsDataFromCSV_InvalidCSV_ReturnsError(t *testing.T) {
	csv := "team_ref,role,user_ref\nt1,member\n"

	got, err := parseTeamsDataFromCSV(bytes.NewBufferString(csv))
	require.Error(t, err)
	require.Nil(t, got)
}

func TestParseTeamsDataByExtension_CSV(t *testing.T) {
	csv := "team_ref,role,user_ref\nt1,member,u1\nt1,owner,u2\n"

	got, err := ParseTeamsDataByExtension(bytes.NewBufferString(csv), "csv")
	require.NoError(t, err)

	require.Len(t, got, 1)
	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{"u1"},
		Owners:      []string{"u2"},
		Visibility:  "private",
	}, got["t1"])
}

func TestParseTeamsDataByExtension_JSON(t *testing.T) {
	json := `{"t1": {"description": "Team 1", "members": ["u1"], "owners": ["u2"], "includeMe": true}}`

	got, err := ParseTeamsDataByExtension(bytes.NewBufferString(json), "json")
	require.NoError(t, err)

	require.Len(t, got, 1)
	teamData := got["t1"]
	require.Equal(t, "Team 1", teamData.Description)
	require.Equal(t, []string{"u1"}, teamData.Members)
	require.Equal(t, []string{"u2"}, teamData.Owners)
	require.True(t, teamData.IncludeMe)
}

func TestParseTeamsDataByExtension_YAML(t *testing.T) {
	yaml := "t1:\n  description: Team 1\n  members:\n    - u1\n  owners:\n    - u2\n  includeMe: true\n"

	got, err := ParseTeamsDataByExtension(bytes.NewBufferString(yaml), "yaml")
	require.NoError(t, err)

	require.Len(t, got, 1)
	teamData := got["t1"]
	require.Equal(t, "Team 1", teamData.Description)
	require.True(t, teamData.IncludeMe)
}

func TestParseTeamsDataByExtension_UnsupportedExtension(t *testing.T) {
	got, err := ParseTeamsDataByExtension(bytes.NewBufferString("data"), "xml")
	require.Error(t, err)
	require.Nil(t, got)
	require.Contains(t, err.Error(), "unsupported file extension")
}

func TestTransformCSVRowsToTeamData_GroupsByTeam(t *testing.T) {
	rows := []map[string]string{
		{"team_ref": "t1", "role": "member", "user_ref": "u1"},
		{"team_ref": "t1", "role": "owner", "user_ref": "u2"},
		{"team_ref": "t2", "role": "member", "user_ref": "u3"},
		{"team_ref": "t3", "role": "owner", "user_ref": "u4"},
		{"team_ref": "t3", "role": "owner", "user_ref": "u5"},
	}

	got := transformCSVRowsToTeamData(rows)

	require.Len(t, got, 3)
	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{"u1"},
		Owners:      []string{"u2"},
		Visibility:  "private",
	}, got["t1"])
	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{"u3"},
		Owners:      []string{},
		Visibility:  "private",
	}, got["t2"])
	require.Equal(t, TeamData{
		Description: "",
		Members:     []string{},
		Owners:      []string{"u4", "u5"},
		Visibility:  "private",
	}, got["t3"])
}
