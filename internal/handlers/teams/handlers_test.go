package teams

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/client"
	corecreator "github.com/pzsp-teams/teams-cli/internal/core/creator"
	teamcreator "github.com/pzsp-teams/teams-cli/internal/teams/creator"
	handlertestutil "github.com/pzsp-teams/teams-cli/internal/testutil/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestArchiveTeam(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantError bool
	}{
		{name: "success"},
		{name: "error", err: errors.New("archive failed"), wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			mockTeams.EXPECT().Archive(gomock.Any(), "team-1", true).Return(tc.err)
			client.SetInstance(&client.Client{Teams: mockTeams})

			var out bytes.Buffer
			result, err := ArchiveTeam(context.Background(), &out, map[string]any{
				"team":          "team-1",
				"spo-read-only": true,
			})
			if tc.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Nil(t, result)
			assert.Contains(t, out.String(), "Team archive initiated")
		})
	}
}

func TestUnarchiveTeam(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantError bool
	}{
		{name: "success"},
		{name: "error", err: errors.New("unarchive failed"), wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			mockTeams.EXPECT().Unarchive(gomock.Any(), "team-1").Return(tc.err)
			client.SetInstance(&client.Client{Teams: mockTeams})

			var out bytes.Buffer
			result, err := UnarchiveTeam(context.Background(), &out, map[string]any{"team": "team-1"})
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Nil(t, result)
			assert.Contains(t, out.String(), "Team unarchive initiated")
		})
	}
}

func TestDeleteTeam(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantError bool
	}{
		{name: "success"},
		{name: "error", err: errors.New("delete failed"), wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			mockTeams.EXPECT().Delete(gomock.Any(), "team-1").Return(tc.err)
			client.SetInstance(&client.Client{Teams: mockTeams})

			var out bytes.Buffer
			result, err := DeleteTeam(context.Background(), &out, map[string]any{"team": "team-1"})
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Nil(t, result)
			assert.Contains(t, out.String(), "Team deletion initiated")
		})
	}
}

func TestGetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	defer client.ResetInstance()

	visibility := "public"
	team := &models.Team{ID: "team-1", DisplayName: "Alpha", Description: "Desc", IsArchived: true, Visibility: &visibility}

	mockTeams := handlertestutil.NewMockTeamClient(ctrl)
	mockTeams.EXPECT().Get(gomock.Any(), "team-1").Return(team, nil)
	client.SetInstance(&client.Client{Teams: mockTeams})

	var out bytes.Buffer
	result, err := GetTeam(context.Background(), &out, map[string]any{"team": "team-1"})
	require.NoError(t, err)
	require.Equal(t, team, result)
	assert.Contains(t, out.String(), "Display Name: Alpha")
	assert.Contains(t, out.String(), "Visibility: public")
}

func TestCreateTeams(t *testing.T) {
	cases := []struct {
		name         string
		dryRun       bool
		expectedLine string
	}{
		{name: "create", dryRun: false, expectedLine: "Team creation completed"},
		{name: "dry run", dryRun: true, expectedLine: "Dry run completed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			dataFile := filepath.Join(t.TempDir(), "teams.yaml")
			data := "engineering:\n  description: \"Engineering\"\n  owners:\n    - owner@example.com\n  members:\n    - member@example.com\n"
			require.NoError(t, os.WriteFile(dataFile, []byte(data), 0o644))

			results := []teamcreator.TeamCreateResult{
				{
					TeamName:    "engineering",
					TeamID:      "team-1",
					Description: "Engineering",
					OwnerRefs:   []string{"owner@example.com"},
					MemberRefs:  []string{"member@example.com"},
					Visibility:  "public",
				},
			}

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			mockTeams.EXPECT().Create(gomock.Any(), gomock.Any(), tc.dryRun).Return(results)
			client.SetInstance(&client.Client{Teams: mockTeams})

			var out bytes.Buffer
			result, err := CreateTeams(context.Background(), &out, map[string]any{
				"data":    dataFile,
				"dry-run": tc.dryRun,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Contains(t, out.String(), tc.expectedLine)
		})
	}
}

func TestCreateSingleTeam(t *testing.T) {
	cases := []struct {
		name      string
		teamName  string
		file      string
		wantError bool
	}{
		{name: "missing team name", teamName: "", file: "", wantError: true},
		{name: "success", teamName: "Alpha", wantError: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			dataFile := tc.file
			if tc.teamName != "" {
				dataFile = filepath.Join(t.TempDir(), "team.yaml")
				data := "owners:\n  - owner@example.com\nmembers:\n  - member@example.com\n"
				require.NoError(t, os.WriteFile(dataFile, []byte(data), 0o644))
			}

			results := []teamcreator.TeamCreateResult{
				{
					TeamName:    "Alpha",
					TeamID:      "team-1",
					Status:      corecreator.StatusCreated,
					Description: "Example team",
					OwnerRefs:   []string{"owner@example.com"},
					MemberRefs:  []string{"member@example.com"},
					Visibility:  "public",
				},
			}

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			if tc.teamName != "" {
				mockTeams.EXPECT().Create(gomock.Any(), gomock.Any(), false).Return(results)
				client.SetInstance(&client.Client{Teams: mockTeams})
			}

			var out bytes.Buffer
			_, err := CreateSingleTeam(context.Background(), &out, map[string]any{
				"team-name":   tc.teamName,
				"description": "Example team",
				"visibility":  "public",
				"file":        dataFile,
				"include-me":  false,
				"dry-run":     false,
			})
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, out.String(), "Created - team: Alpha")
		})
	}
}
