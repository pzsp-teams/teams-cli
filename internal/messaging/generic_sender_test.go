package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/models"
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

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if sendCalled {
		t.Error("SendMessage should not be called in dry run mode")
	}

	if results[0].Err != nil {
		t.Errorf("unexpected error in dry run: %v", results[0].Err)
	}
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Err == nil {
		t.Error("expected error for mention resolution failure")
	}

	if !errors.Is(results[0].Err, ErrMentionResolutionFailed) {
		t.Errorf("expected ErrMentionResolutionFailed, got: %v", results[0].Err)
	}
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

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Err == nil {
		t.Error("expected error for send failure")
	}

	if !errors.Is(results[0].Err, ErrMessageSendFailed) {
		t.Errorf("expected ErrMessageSendFailed, got: %v", results[0].Err)
	}
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

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if callCount != 1 {
		t.Errorf("expected exactly 1 send call before stopping, got %d", callCount)
	}

	errorCount := 0
	for _, result := range results {
		if result.Err != nil {
			errorCount++
		}
	}

	if errorCount != 3 {
		t.Errorf("expected 3 errors (1 fail + 2 skipped), got %d", errorCount)
	}

	skippedCount := 0
	failedCount := 0
	for _, result := range results {
		if errors.Is(result.Err, ErrMessageSkipped) {
			skippedCount++
		} else if result.getRef() == failedRef {
			failedCount++
		}
	}

	if failedCount != 1 {
		t.Errorf("expected 1 failed message, got %d", failedCount)
	}

	if skippedCount != 2 {
		t.Errorf("expected 2 skipped messages, got %d", skippedCount)
	}
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

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if callCount != 3 {
		t.Errorf("expected 3 send calls (ignoring errors), got %d", callCount)
	}

	successCount := 0
	for _, result := range results {
		if result.Err == nil {
			successCount++
		}
	}

	if successCount != 2 {
		t.Errorf("expected 2 successful sends, got %d", successCount)
	}
}

func TestGenericSender_ProcessMentions_NoMentions(t *testing.T) {
	adapter := &mockAdapter{}
	sender := newGenericSender(adapter, newMockResult, senderTypeChannel)

	content, mentions, err := sender.processMentions(context.Background(), "ref1", "message without mentions")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if content != "message without mentions" {
		t.Errorf("expected unchanged content, got: %s", content)
	}

	if len(mentions) != 0 {
		t.Errorf("expected no mentions, got %d", len(mentions))
	}
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mentions) != 1 {
		t.Errorf("expected 1 mention, got %d", len(mentions))
	}

	if content == "Hello @@user@@" {
		t.Error("expected mention to be replaced in content")
	}
}
