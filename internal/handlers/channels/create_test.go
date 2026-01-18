package channels

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	channelcreator "github.com/pzsp-teams/teams-cli/internal/channels/creator"
	"github.com/pzsp-teams/teams-cli/internal/client"
	corecreator "github.com/pzsp-teams/teams-cli/internal/core/creator"
	handlertestutil "github.com/pzsp-teams/teams-cli/internal/testutil/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateChannels(t *testing.T) {
	cases := []struct {
		name         string
		dryRun       bool
		status       channelcreator.Status
		expectedLine string
	}{
		{
			name:         "create",
			dryRun:       false,
			status:       corecreator.StatusCreated,
			expectedLine: "Channel creation completed",
		},
		{
			name:         "dry run",
			dryRun:       true,
			status:       corecreator.StatusWouldCreate,
			expectedLine: "Dry run completed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			dataFile := filepath.Join(t.TempDir(), "channels.yaml")
			data := "general:\n  owners:\n    - owner@example.com\n  members:\n    - member@example.com\n"
			require.NoError(t, os.WriteFile(dataFile, []byte(data), 0o644))

			results := []channelcreator.CreateResult{
				{
					ChannelName: "general",
					Status:      tc.status,
					OwnerRefs:   []string{"owner@example.com"},
					MemberRefs:  []string{"member@example.com"},
				},
			}

			mockChannels := handlertestutil.NewMockChannelClient(ctrl)
			mockChannels.EXPECT().Create(gomock.Any(), "team-1", gomock.Any(), false, false, tc.dryRun).Return(results)
			client.SetInstance(&client.Client{Channels: mockChannels})

			var out bytes.Buffer
			result, err := CreateChannels(context.Background(), &out, map[string]any{
				"team":               "team-1",
				"data":               dataFile,
				"dry-run":            tc.dryRun,
				"ensure-in-channels": false,
				"ensure-in-team":     false,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Contains(t, out.String(), tc.expectedLine)
		})
	}
}
