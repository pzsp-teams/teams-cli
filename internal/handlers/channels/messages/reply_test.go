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

func TestReplyToMessage(t *testing.T) {
	cases := []struct {
		name      string
		result    sender.ChannelSendResult
		wantError bool
	}{
		{
			name:   "success",
			result: sender.ChannelSendResult{ChannelRef: "general"},
		},
		{
			name:      "error",
			result:    sender.ChannelSendResult{Error: errors.New("reply failed")},
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockChannels := handlertestutil.NewMockChannelClient(ctrl)
			mockChannels.EXPECT().SendReply(gomock.Any(), "team-1", "general", "msg-1", gomock.Any()).Return(tc.result)
			client.SetInstance(&client.Client{Channels: mockChannels})

			var out bytes.Buffer
			_, err := ReplyToMessage(context.Background(), &out, map[string]any{
				"team":         "team-1",
				"channel":      "general",
				"message-id":   "msg-1",
				"message":      "Thanks",
				"message-file": "",
			})
			if tc.wantError {
				require.Error(t, err)
				assert.Contains(t, out.String(), "Failed to send reply")
				return
			}
			require.NoError(t, err)
			assert.Contains(t, out.String(), "Reply sent successfully")
		})
	}
}
