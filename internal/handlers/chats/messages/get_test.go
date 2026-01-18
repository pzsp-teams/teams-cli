package messages

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/chats/retriever"
	"github.com/pzsp-teams/teams-cli/internal/client"
	handlertestutil "github.com/pzsp-teams/teams-cli/internal/testutil/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	defer client.ResetInstance()

	createdAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
	messages := []*retriever.ChatMessageWithContext{
		{
			ChatName: "Project",
			ChatType: "group",
			Message: &models.Message{
				ID:              "msg-1",
				Content:         "Hi",
				CreatedDateTime: createdAt,
				From:            &models.MessageFrom{DisplayName: "Ada"},
			},
		},
	}

	mockChats := handlertestutil.NewMockChatClient(ctrl)
	mockChats.EXPECT().GetMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return(messages, nil)
	client.SetInstance(&client.Client{Chats: mockChats})

	var out bytes.Buffer
	result, err := GetMessages(context.Background(), &out, map[string]any{
		"start":    "2024-01-02 12:00",
		"end":      "2024-01-02 13:00",
		"format":   "plain",
		"chat-ref": "chat-1",
	})
	require.NoError(t, err)
	require.Len(t, result.([]*retriever.ChatMessageWithContext), 1)
	assert.Contains(t, out.String(), "CHAT:")
	assert.Contains(t, out.String(), "Project")
	assert.Contains(t, out.String(), "Hi")
}
