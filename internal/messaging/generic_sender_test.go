package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResult struct {
	Ref     string
	Message string
	Err     error
}

func (m mockResult) getError() error {
	return m.Err
}

func (m mockResult) getRef() string {
	return m.Ref
}

func (m mockResult) getMessage() string {
	return m.Message
}

type mockAdapter struct {
	getMentionsFn func(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error)
	sendMessageFn func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error)
}

func (m *mockAdapter) getMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
	if m.getMentionsFn != nil {
		return m.getMentionsFn(ctx, ref, rawMentions)
	}
	return nil, nil
}

func (m *mockAdapter) sendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
	if m.sendMessageFn != nil {
		return m.sendMessageFn(ctx, ref, body)
	}
	return &models.Message{Content: body.Content}, nil
}

func newMockResult(ref, message string, err error) mockResult {
	return mockResult{Ref: ref, Message: message, Err: err}
}

func TestGenericSender_Send_Success(t *testing.T) {
	adapter := &mockAdapter{
		sendMessageFn: func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message1",
		"ref2": "message2",
	}

	results := sender.send(context.Background(), messages, false, false)

	require.Len(t, results, 2)
	for _, result := range results {
		assert.NoError(t, result.Err)
	}
}

func TestGenericSender_Send_DryRun(t *testing.T) {
	sendCalled := false
	adapter := &mockAdapter{
		sendMessageFn: func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
			sendCalled = true
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message1",
	}

	results := sender.send(context.Background(), messages, true, false)

	require.Len(t, results, 1)
	assert.False(t, sendCalled, "SendMessage should not be called in dry run mode")
	assert.NoError(t, results[0].Err)
}

func TestGenericSender_Send_MentionError(t *testing.T) {
	mentionErr := errors.New("mention error")
	adapter := &mockAdapter{
		getMentionsFn: func(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
			return nil, mentionErr
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message with @@mention@@",
	}

	results := sender.send(context.Background(), messages, false, false)

	require.Len(t, results, 1)
	require.Error(t, results[0].Err)
	require.ErrorIs(t, results[0].Err, ErrMentionResolutionFailed)
}

func TestGenericSender_Send_SendError(t *testing.T) {
	sendErr := errors.New("send error")
	adapter := &mockAdapter{
		sendMessageFn: func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
			return nil, sendErr
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message1",
	}

	results := sender.send(context.Background(), messages, false, false)

	require.Len(t, results, 1)
	require.Error(t, results[0].Err)
	require.ErrorIs(t, results[0].Err, ErrMessageSendFailed)
}

func TestGenericSender_Send_StopOnError(t *testing.T) {
	callCount := 0
	failedRef := ""
	adapter := &mockAdapter{
		sendMessageFn: func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
			callCount++
			if callCount == 1 {
				failedRef = ref
				return nil, errors.New("send error")
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message1",
		"ref2": "message2",
		"ref3": "message3",
	}

	results := sender.send(context.Background(), messages, false, false)

	require.Len(t, results, 3)
	assert.Equal(t, 1, callCount, "expected exactly 1 send call before stopping")

	errorCount := 0
	for _, result := range results {
		if result.Err != nil {
			errorCount++
		}
	}
	assert.Equal(t, 3, errorCount, "expected 3 errors (1 fail + 2 skipped)")

	skippedCount := 0
	failedCount := 0
	for _, result := range results {
		if errors.Is(result.Err, ErrMessageSkipped) {
			skippedCount++
		} else if result.getRef() == failedRef {
			failedCount++
		}
	}

	assert.Equal(t, 1, failedCount)
	assert.Equal(t, 2, skippedCount)
}

func TestGenericSender_Send_IgnoreErrors(t *testing.T) {
	callCount := 0
	adapter := &mockAdapter{
		sendMessageFn: func(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
			callCount++
			if ref == "ref1" {
				return nil, errors.New("send error")
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	messages := map[string]string{
		"ref1": "message1",
		"ref2": "message2",
		"ref3": "message3",
	}

	results := sender.send(context.Background(), messages, false, true)

	require.Len(t, results, 3)
	assert.Equal(t, 3, callCount, "expected 3 send calls (ignoring errors)")

	successCount := 0
	for _, result := range results {
		if result.Err == nil {
			successCount++
		}
	}

	assert.Equal(t, 2, successCount)
}

func TestGenericSender_ProcessMentions_NoMentions(t *testing.T) {
	adapter := &mockAdapter{}
	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	content, mentions, err := sender.processMentions(context.Background(), "ref1", "message without mentions")

	assert.NoError(t, err)
	assert.Equal(t, "message without mentions", content)
	assert.Empty(t, mentions)
}

func TestGenericSender_ProcessMentions_WithMentions(t *testing.T) {
	adapter := &mockAdapter{
		getMentionsFn: func(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
			return []models.Mention{
				{
					TargetID: "user-id",
					Text:     "User Name",
				},
			}, nil
		},
	}

	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	content, mentions, err := sender.processMentions(context.Background(), "ref1", "Hello @@user@@")

	assert.NoError(t, err)
	assert.Len(t, mentions, 1)
	assert.NotEqual(t, "Hello @@user@@", content, "expected mention to be replaced in content")
}
