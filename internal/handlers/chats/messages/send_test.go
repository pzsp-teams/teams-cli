package messages

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/teams-cli/internal/chats/sender"
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
		results      []sender.ChatSendResult
		expectedLine string
	}{
		{
			name:         "send",
			dryRun:       false,
			results:      []sender.ChatSendResult{{ChatRef: "chat-1", Message: "Hello"}},
			expectedLine: "Send complete - successful: 1, total: 1",
		},
		{
			name:         "dry run",
			dryRun:       true,
			results:      []sender.ChatSendResult{{ChatRef: "chat-1", Message: "Hello"}},
			expectedLine: "Would send - chat: chat-1",
		},
		{
			name:         "error",
			dryRun:       false,
			results:      []sender.ChatSendResult{{ChatRef: "chat-1", Error: errors.New("send failed")}},
			expectedLine: "Failed - chat: chat-1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockChats := handlertestutil.NewMockChatClient(ctrl)
			mockChats.EXPECT().Send(gomock.Any(), gomock.Any(), tc.dryRun, false).Return(tc.results)
			client.SetInstance(&client.Client{Chats: mockChats})

			var out bytes.Buffer
			result, err := SendMessages(context.Background(), &out, map[string]any{
				"message":       "Hello",
				"chats":         []string{"chat-1"},
				"dry-run":       tc.dryRun,
				"ignore-errors": false,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Contains(t, out.String(), tc.expectedLine)
		})
	}
}
