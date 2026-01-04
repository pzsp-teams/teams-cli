package retriever

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/cli/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetMessages_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// Expected messages from the service
	expectedMessages := []*models.Message{
		{
			ID:              "msg1",
			Content:         "<p>Hello</p>",
			ContentType:     models.MessageContentTypeHTML,
			CreatedDateTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:              "msg2",
			Content:         "Plain text message",
			ContentType:     models.MessageContentTypeText,
			CreatedDateTime: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
	}

	top := int32(50)
	mockService.EXPECT().
		ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top).
		Return(expectedMessages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)

	assert.Equal(t, "msg1", messages[0].ID)
	assert.Equal(t, "Hello\n\n", messages[0].Content)

	assert.Equal(t, "msg2", messages[1].ID)
	assert.Equal(t, "Plain text message", messages[1].Content)
}

func TestGetMessages_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	top := int32(50)
	mockService.EXPECT().
		ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top).
		Return([]*models.Message{}, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Empty(t, messages)
}

func TestGetMessages_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	top := int32(50)
	expectedError := errors.New("API error")
	mockService.EXPECT().
		ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top).
		Return(nil, expectedError)

	retriever := NewRetriever(mockService, formatterInstance)
	messages, err := retriever.GetMessages(ctx, timeRange)

	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.ErrorIs(t, err, ErrListingMessagesFailed)
	assert.Contains(t, err.Error(), "API error")
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

	// Mix of HTML and plain text messages
	messages := []*models.Message{
		{
			ID:          "msg1",
			Content:     "<p>HTML</p>",
			ContentType: models.MessageContentTypeHTML,
		},
		{
			ID:          "msg2",
			Content:     "Plain",
			ContentType: models.MessageContentTypeText,
		},
		{
			ID:          "msg3",
			Content:     "<b>More HTML</b>",
			ContentType: models.MessageContentTypeHTML,
		},
	}

	top := int32(50)
	mockService.EXPECT().
		ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top).
		Return(messages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	assert.Equal(t, "HTML\n\n", result[0].Content)  // HTML - formatted
	assert.Equal(t, "Plain", result[1].Content)     // Text - not formatted
	assert.Equal(t, "More HTML", result[2].Content) // HTML - formatted
}

func TestGetMessages_ComplexHTML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := testutil.NewMockChatsService(ctrl)
	formatterInstance := formatter.NewPlainTextFormatter()

	ctx := context.Background()
	timeRange := coreretriever.TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	messages := []*models.Message{
		{
			ID:          "msg1",
			Content:     "Hello <at id=\"0\">User</at>!<br>Check <a href=\"https://example.com\">this link</a>",
			ContentType: models.MessageContentTypeHTML,
		},
	}

	top := int32(50)
	mockService.EXPECT().
		ListAllMessages(ctx, &timeRange.Start, &timeRange.End, &top).
		Return(messages, nil)

	retriever := NewRetriever(mockService, formatterInstance)
	result, err := retriever.GetMessages(ctx, timeRange)

	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// Verify the HTML was properly formatted to plain text
	expected := "Hello @User!\nCheck this link (https://example.com)"
	assert.Equal(t, expected, result[0].Content)
}
