package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/channels"
	"github.com/pzsp-teams/lib/models"
)

type channelServiceStub struct {
	channels.Service

	getMentionsFunc func(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error)
	sendMessageFunc func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error)
}

func (m *channelServiceStub) GetMentions(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
	if m.getMentionsFunc != nil {
		return m.getMentionsFunc(ctx, teamRef, channelRef, rawMentions)
	}
	return nil, nil
}

func (m *channelServiceStub) SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, teamRef, channelRef, body)
	}
	return &models.Message{Content: body.Content}, nil
}

func TestSendToChannels_SuccessfulSend(t *testing.T) {
	stub := &channelServiceStub{
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello World",
		"channel2": "Test Message",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Expected no error for %s, got %v", result.ChannelRef, result.Error)
		}
	}
}

func TestSendToChannels_DryRun(t *testing.T) {
	stub := &channelServiceStub{
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			t.Error("SendMessage should not be called during dry run")
			return nil, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello World",
	}

	results := sender.Send(context.Background(), "team1", messages, true, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Errorf("Expected no error in dry run, got %v", results[0].Error)
	}

	if results[0].Message != "Hello World" {
		t.Errorf("Expected message 'Hello World', got %q", results[0].Message)
	}
}

func TestSendToChannels_WithMentions(t *testing.T) {
	stub := &channelServiceStub{
		getMentionsFunc: func(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
			if len(rawMentions) != 1 || rawMentions[0] != "alice" {
				t.Errorf("Expected rawMentions [alice], got %v", rawMentions)
			}
			return []models.Mention{
				{
					Kind:     models.MentionUser,
					AtID:     0,
					Text:     "Alice Smith",
					TargetID: "user123",
				},
			}, nil
		},
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			expectedContent := `Hello <at id="0">Alice Smith</at>!`
			if body.Content != expectedContent {
				t.Errorf("Expected content %q, got %q", expectedContent, body.Content)
			}
			if len(body.Mentions) != 1 {
				t.Errorf("Expected 1 mention, got %d", len(body.Mentions))
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello @@alice@@!",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Errorf("Expected no error, got %v", results[0].Error)
	}
}

func TestSendToChannels_MentionResolutionError(t *testing.T) {
	mentionErr := errors.New("user not found")
	stub := &channelServiceStub{
		getMentionsFunc: func(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
			return nil, mentionErr
		},
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			t.Error("SendMessage should not be called when mention resolution fails")
			return nil, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello @@alice@@!",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected error for mention resolution failure")
	}

	if results[0].ChannelRef != "channel1" {
		t.Errorf("Expected ChannelRef 'channel1', got %q", results[0].ChannelRef)
	}
}

func TestSendToChannels_SendError(t *testing.T) {
	sendErr := errors.New("network error")
	stub := &channelServiceStub{
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			return nil, sendErr
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello World",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected error for send failure")
	}
}

func TestSendToChannels_StopOnError(t *testing.T) {
	sendErr := errors.New("network error")
	callCount := 0
	failedRef := ""

	stub := &channelServiceStub{
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			callCount++
			if callCount == 1 {
				failedRef = channelRef
				return nil, sendErr
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Message 1",
		"channel2": "Message 2",
		"channel3": "Message 3",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if callCount != 1 {
		t.Errorf("Expected SendMessage to be called 1 time, got %d", callCount)
	}

	errorCount := 0
	skippedCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Error != nil {
			switch {
			case errors.Is(result.Error, ErrMessageSkipped):
				skippedCount++
			case result.ChannelRef == failedRef:
				failedCount++
			default:
				errorCount++
			}
		}
	}

	if failedCount != 1 {
		t.Errorf("Expected 1 failed message, got %d", failedCount)
	}

	if errorCount != 0 {
		t.Errorf("Expected 0 other errors, got %d", errorCount)
	}

	if skippedCount != 2 {
		t.Errorf("Expected 2 skipped messages, got %d", skippedCount)
	}
}

func TestSendToChannels_IgnoreErrors(t *testing.T) {
	sendErr := errors.New("network error")
	callCount := 0

	stub := &channelServiceStub{
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			callCount++
			if channelRef == "channel1" {
				return nil, sendErr
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Message 1",
		"channel2": "Message 2",
		"channel3": "Message 3",
	}

	results := sender.Send(context.Background(), "team1", messages, false, true)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if callCount != 3 {
		t.Errorf("Expected SendMessage to be called 3 times, got %d", callCount)
	}

	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	if errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", errorCount)
	}

	if successCount != 2 {
		t.Errorf("Expected 2 successful sends, got %d", successCount)
	}
}

func TestSendToChannels_DuplicateMentions(t *testing.T) {
	stub := &channelServiceStub{
		getMentionsFunc: func(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
			if len(rawMentions) != 2 {
				t.Errorf("Expected 2 raw mentions (including duplicate), got %d", len(rawMentions))
			}
			if rawMentions[0] != "alice" || rawMentions[1] != "alice" {
				t.Errorf("Expected [alice, alice], got %v", rawMentions)
			}

			return []models.Mention{
				{Kind: models.MentionUser, AtID: 0, Text: "Alice Smith", TargetID: "user123"},
				{Kind: models.MentionUser, AtID: 1, Text: "Alice Smith", TargetID: "user123"},
			}, nil
		},
		sendMessageFunc: func(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error) {
			expectedContent := `Hello <at id="0">Alice Smith</at>, this is for <at id="1">Alice Smith</at> again`
			if body.Content != expectedContent {
				t.Errorf("Expected content %q, got %q", expectedContent, body.Content)
			}
			if len(body.Mentions) != 2 {
				t.Errorf("Expected 2 mentions, got %d", len(body.Mentions))
			}
			return &models.Message{Content: body.Content}, nil
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello @@alice@@, this is for @@alice@@ again",
	}

	results := sender.Send(context.Background(), "team1", messages, false, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Errorf("Expected no error, got %v", results[0].Error)
	}
}

func TestSendToChannels_DryRunWithMentionError(t *testing.T) {
	mentionErr := errors.New("user not found")
	stub := &channelServiceStub{
		getMentionsFunc: func(ctx context.Context, teamRef, channelRef string, rawMentions []string) ([]models.Mention, error) {
			return nil, mentionErr
		},
	}

	sender := NewChannelSender(stub)
	messages := map[string]string{
		"channel1": "Hello @@alice@@!",
	}

	results := sender.Send(context.Background(), "team1", messages, true, false)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Error("Expected error for mention resolution failure in dry run")
	}
}
