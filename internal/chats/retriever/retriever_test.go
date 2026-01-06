package retriever

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func stringPtr(s string) *string {
	return &s
}

func TestGetMessages_Success_BothChatTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	oneOnOneTopic := "One on One Chat"
	groupTopic := "Group Chat"
	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: &oneOnOneTopic,
			Type:  models.ChatTypeOneOnOne,
		},
		{
			ID:    "chat2-id",
			Topic: &groupTopic,
			Type:  models.ChatTypeGroup,
		},
	}

	messagesChat1 := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg1",
				Content:         "<p>Hello</p>",
				ContentType:     models.MessageContentTypeHTML,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				From: &models.MessageFrom{
					DisplayName: "User 1",
					UserID:      "user1",
				},
			},
		},
		NextLink: nil,
	}

	messagesChat2 := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg2",
				Content:         "Plain text message",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
				From: &models.MessageFrom{
					DisplayName: "User 2",
					UserID:      "user2",
				},
			},
		},
		NextLink: nil,
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.OneOnOneChatRef{Ref: "chat1-id"}, false, nil).
		Return(messagesChat1, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat2-id"}, false, nil).
		Return(messagesChat2, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)

	assert.Equal(t, "msg1", messages[0].Message.ID)
	assert.Equal(t, "Hello", messages[0].Message.Content)
	assert.Equal(t, "One on One Chat", messages[0].ChatName)
	assert.Equal(t, "chat1-id", messages[0].ChatID)
	assert.Equal(t, "one-on-one", messages[0].ChatType)

	assert.Equal(t, "msg2", messages[1].Message.ID)
	assert.Equal(t, "Plain text message", messages[1].Message.Content)
	assert.Equal(t, "Group Chat", messages[1].ChatName)
	assert.Equal(t, "chat2-id", messages[1].ChatID)
	assert.Equal(t, "group", messages[1].ChatType)
}

func TestGetMessages_TimeRangeFiltering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: stringPtr("Test Chat"),
			Type:  models.ChatTypeGroup,
		},
	}

	messages := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg-before",
				Content:         "Before range",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			{
				ID:              "msg-inside",
				Content:         "Inside range",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:              "msg-after",
				Content:         "After range",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			},
		},
		NextLink: nil,
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat1-id"}, false, nil).
		Return(messages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "msg-inside", result[0].Message.ID)
}

func TestGetMessages_NoChatsFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return([]*models.Chat{}, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.ErrorIs(t, err, ErrNoChatsFound)
}

func TestGetMessages_ListChatsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	expectedError := errors.New("API error")
	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(nil, expectedError)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.ErrorIs(t, err, ErrListingChatsFailed)
	assert.Contains(t, err.Error(), "API error")
}

func TestGetMessages_ListMessagesFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: stringPtr("Test Chat"),
			Type:  models.ChatTypeGroup,
		},
	}

	expectedError := errors.New("API error")
	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat1-id"}, false, nil).
		Return(nil, expectedError)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.ErrorIs(t, err, ErrListingMessagesFailed)
	assert.Contains(t, err.Error(), "API error")
}

func TestGetMessages_403ErrorIgnored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: stringPtr("Accessible Chat"),
			Type:  models.ChatTypeGroup,
		},
		{
			ID:    "chat2-id",
			Topic: stringPtr("Forbidden Chat"),
			Type:  models.ChatTypeGroup,
		},
	}

	messagesChat1 := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg1",
				Content:         "Accessible message",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		NextLink: nil,
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat1-id"}, false, nil).
		Return(messagesChat1, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat2-id"}, false, nil).
		Return(nil, fmt.Errorf("403 Forbidden"))

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "msg1", messages[0].Message.ID)
}

func TestGetMessages_OnlyHTMLMessagesFormatted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: stringPtr("Test Chat"),
			Type:  models.ChatTypeGroup,
		},
	}

	messages := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg1",
				Content:         "<p>HTML</p>",
				ContentType:     models.MessageContentTypeHTML,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:              "msg2",
				Content:         "Plain",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:              "msg3",
				Content:         "<b>More HTML</b>",
				ContentType:     models.MessageContentTypeHTML,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		NextLink: nil,
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat1-id"}, false, nil).
		Return(messages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	assert.Equal(t, "HTML", result[0].Message.Content)
	assert.Equal(t, "Plain", result[1].Message.Content)
	assert.Equal(t, "More HTML", result[2].Message.Content)
}

func TestGetMessages_ChatWithoutTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: nil,
			Type:  models.ChatTypeOneOnOne,
		},
	}

	messages := &models.MessageCollection{
		Messages: []*models.Message{
			{
				ID:              "msg1",
				Content:         "Test message",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		NextLink: nil,
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.OneOnOneChatRef{Ref: "chat1-id"}, false, nil).
		Return(messages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "chat1-id", result[0].ChatName)
}

func TestGetMessages_EmptyMessagesInChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	chatList := []*models.Chat{
		{
			ID:    "chat1-id",
			Topic: stringPtr("Empty Chat"),
			Type:  models.ChatTypeGroup,
		},
	}

	mockService.EXPECT().
		ListChats(ctx, nil).
		Return(chatList, nil)

	mockService.EXPECT().
		ListMessages(ctx, chats.GroupChatRef{Ref: "chat1-id"}, false, nil).
		Return(&models.MessageCollection{Messages: []*models.Message{}, NextLink: nil}, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Empty(t, messages)
}
