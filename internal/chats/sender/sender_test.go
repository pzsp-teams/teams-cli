package sender

import (
	"context"
	"errors"
	"testing"

	coresender "github.com/pzsp-teams/cli/internal/core/sender"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSendToChats_SuccessfulSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, chatRef chats.ChatRef, body models.MessageBody) (*models.Message, error) {
			return &models.Message{Content: body.Content}, nil
		}).
		Times(2)

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Hello World",
		"chat2": "Test Message",
	}

	results := sender.Send(context.Background(), messages, false, false)

	require.Len(t, results, 2)
	for _, result := range results {
		assert.NoError(t, result.Error)
	}
}

func TestSendToChats_DryRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0) // Should not be called during dry run

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Hello World",
	}

	results := sender.Send(context.Background(), messages, true, false)

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
	assert.Equal(t, "Hello World", results[0].Message)
}

func TestSendToChats_WithMentions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		GetMentions(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, chatRef chats.ChatRef, rawMentions []string) ([]models.Mention, error) {
			require.Len(t, rawMentions, 1)
			require.Equal(t, "alice", rawMentions[0])
			return []models.Mention{
				{
					Kind:     models.MentionUser,
					AtID:     0,
					Text:     "Alice Smith",
					TargetID: "user123",
				},
			}, nil
		}).
		Times(1)

	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, chatRef chats.ChatRef, body models.MessageBody) (*models.Message, error) {
			expectedContent := `Hello <at id="0">Alice Smith</at>!`
			require.Equal(t, expectedContent, body.Content)
			require.Len(t, body.Mentions, 1)
			return &models.Message{Content: body.Content}, nil
		}).
		Times(1)

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Hello @@alice@@!",
	}

	results := sender.Send(context.Background(), messages, false, false)

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Error)
}

func TestSendToChats_MentionResolutionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mentionErr := errors.New("user not found")
	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		GetMentions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, mentionErr).
		Times(1)

	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0) // Should not be called when mention resolution fails

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Hello @@alice@@!",
	}

	results := sender.Send(context.Background(), messages, false, false)

	require.Len(t, results, 1)
	require.Error(t, results[0].Error)
	assert.Equal(t, "chat1", results[0].ChatRef)
}

func TestSendToChats_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sendErr := errors.New("network error")
	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, sendErr).
		Times(1)

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Hello World",
	}

	results := sender.Send(context.Background(), messages, false, false)

	require.Len(t, results, 1)
	require.Error(t, results[0].Error)
}

func TestSendToChats_StopOnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sendErr := errors.New("network error")
	callCount := 0
	failedRef := ""

	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, chatRef chats.ChatRef, body models.MessageBody) (*models.Message, error) {
			callCount++
			if callCount == 1 {
				failedRef = "chat1" // First call fails
				return nil, sendErr
			}
			return &models.Message{Content: body.Content}, nil
		}).
		Times(1) // Should stop after first error

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Message 1",
		"chat2": "Message 2",
		"chat3": "Message 3",
	}

	results := sender.Send(context.Background(), messages, false, false)

	require.Len(t, results, 3)
	assert.Equal(t, 1, callCount)

	errorCount := 0
	skippedCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Error != nil {
			switch {
			case errors.Is(result.Error, coresender.ErrMessageSkipped):
				skippedCount++
			case result.ChatRef == failedRef:
				failedCount++
			default:
				errorCount++
			}
		}
	}

	assert.Equal(t, 1, failedCount)
	assert.Equal(t, 0, errorCount)
	assert.Equal(t, 2, skippedCount)
}

func TestSendToChats_IgnoreErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sendErr := errors.New("network error")
	callCount := 0

	mockSvc := testutil.NewMockChatsService(ctrl)
	mockSvc.EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, chatRef chats.ChatRef, body models.MessageBody) (*models.Message, error) {
			callCount++
			// Fail for chat1, succeed for others
			if callCount == 1 {
				return nil, sendErr
			}
			return &models.Message{Content: body.Content}, nil
		}).
		Times(3) // Should call all 3 times even with error

	sender := NewChatSender(mockSvc)
	messages := map[string]string{
		"chat1": "Message 1",
		"chat2": "Message 2",
		"chat3": "Message 3",
	}

	results := sender.Send(context.Background(), messages, false, true)

	require.Len(t, results, 3)
	assert.Equal(t, 3, callCount)

	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	assert.Equal(t, 1, errorCount)
	assert.Equal(t, 2, successCount)
}
