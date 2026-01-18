package messages

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/channels/retriever"
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

	createdAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	messages := []*retriever.ChannelMessageWithContext{
		{
			TeamName:    "Team A",
			ChannelName: "General",
			Message: &models.Message{
				ID:              "msg-1",
				Content:         "Hello",
				CreatedDateTime: createdAt,
				From:            &models.MessageFrom{DisplayName: "Ada"},
			},
		},
	}

	mockChannels := handlertestutil.NewMockChannelClient(ctrl)
	mockChannels.EXPECT().GetMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(messages, nil)
	client.SetInstance(&client.Client{Channels: mockChannels})

	var out bytes.Buffer
	result, err := GetMessages(context.Background(), &out, map[string]any{
		"start":       "2024-01-01 10:00",
		"end":         "2024-01-01 11:00",
		"format":      "plain",
		"team-ref":    "team-1",
		"channel-ref": "general",
	})
	require.NoError(t, err)
	require.Len(t, result.([]*retriever.ChannelMessageWithContext), 1)
	assert.Contains(t, out.String(), "FROM:")
	assert.Contains(t, out.String(), "Team A")
	assert.Contains(t, out.String(), "General")
	assert.Contains(t, out.String(), "Hello")
}
