package chats

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

func TestListChats(t *testing.T) {
	chatTopic := "Planning"
	cases := []struct {
		name        string
		chats       []*models.Chat
		expectedOut []string
		wantNil     bool
	}{
		{
			name:        "no chats",
			chats:       nil,
			expectedOut: []string{"No chats found"},
			wantNil:     true,
		},
		{
			name: "with chats",
			chats: []*models.Chat{
				{ID: "chat-1", Type: models.ChatTypeOneOnOne},
				{ID: "chat-2", Type: models.ChatTypeGroup, Topic: &chatTopic},
			},
			expectedOut: []string{"Found 2 chats", "ID: chat-1", "Topic: Planning"},
			wantNil:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			defer client.ResetInstance()

			mockChats := handlertestutil.NewMockChatClient(ctrl)
			mockChats.EXPECT().List(gomock.Any()).Return(tc.chats, nil)
			client.SetInstance(&client.Client{Chats: mockChats})

			var out bytes.Buffer
			result, err := ListChats(context.Background(), &out, map[string]any{})
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
