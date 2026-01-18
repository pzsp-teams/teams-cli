package teams

import (
	"bytes"
	"context"
	"testing"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/client"
	handlertestutil "github.com/pzsp-teams/teams-cli/internal/testutil/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestListTeams(t *testing.T) {
	cases := []struct {
		name        string
		teams       []*models.Team
		expectedOut []string
		wantNil     bool
	}{
		{
			name:        "no teams",
			teams:       nil,
			expectedOut: []string{"No teams found."},
			wantNil:     true,
		},
		{
			name: "with teams",
			teams: []*models.Team{
				{ID: "team-1", DisplayName: "Alpha"},
				{ID: "team-2", DisplayName: "Beta", IsArchived: true},
			},
			expectedOut: []string{"Teams:", "Alpha", "Beta (Archived)"},
			wantNil:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockTeams := handlertestutil.NewMockTeamClient(ctrl)
			mockTeams.EXPECT().ListMyJoined(gomock.Any()).Return(tc.teams, nil)
			client.SetInstance(&client.Client{Teams: mockTeams})

			var out bytes.Buffer
			result, err := ListTeams(context.Background(), &out, map[string]any{})
			require.NoError(t, err)
			if tc.wantNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}

			output := out.String()
			for _, expected := range tc.expectedOut {
				assert.Contains(t, output, expected)
			}
		})
	}
}
