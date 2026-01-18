package messages

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/teams-cli/internal/channels/sender"
	"github.com/pzsp-teams/teams-cli/internal/client"
	handlertestutil "github.com/pzsp-teams/teams-cli/internal/testutil/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSendMessages(t *testing.T) {
	cases := []struct {
		name         string
		dryRun       bool
		results      []sender.ChannelSendResult
		expectedLine string
	}{
		{
			name:         "send",
			dryRun:       false,
			results:      []sender.ChannelSendResult{{ChannelRef: "general", Message: "Hello"}},
			expectedLine: "Send complete - successful: 1, total: 1",
		},
		{
			name:         "dry run",
			dryRun:       true,
			results:      []sender.ChannelSendResult{{ChannelRef: "general", Message: "Hello"}},
			expectedLine: "Would send - channel: general",
		},
		{
			name:         "error",
			dryRun:       false,
			results:      []sender.ChannelSendResult{{ChannelRef: "general", Error: errors.New("send failed")}},
			expectedLine: "Failed - channel: general",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockChannels := handlertestutil.NewMockChannelClient(ctrl)
			mockChannels.EXPECT().Send(gomock.Any(), "team-1", gomock.Any(), tc.dryRun, false).Return(tc.results)
			client.SetInstance(&client.Client{Channels: mockChannels})

			var out bytes.Buffer
			result, err := SendMessages(context.Background(), &out, map[string]any{
				"team":          "team-1",
				"message":       "Hello",
				"channels":      []string{"general"},
				"dry-run":       tc.dryRun,
				"ignore-errors": false,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Contains(t, out.String(), tc.expectedLine)
		})
	}
}
