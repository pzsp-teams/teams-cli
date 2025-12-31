package messaging

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/templates"
	"github.com/pzsp-teams/cli/internal/utils"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

type chatSender struct {
	chatService chats.Service
}

// ChatSender defines the interface for sending messages to chats
type ChatSender interface {
	SendToChats(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult
}

// NewChatSender creates a new ChatSender with the provided chats service
func NewChatSender(chatService chats.Service) *chatSender {
	return &chatSender{
		chatService: chatService,
	}
}

// SendToChats sends messages to multiple chats
//
// messages: map of chat reference (name or ID) to message content
// Returns a slice of ChatSendResult containing the outcome for each chat
func (s *chatSender) SendToChats(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult {
	logger := initializers.Logger.With("dryRun", dryRun, "ignoreError", ignoreError, "total", len(messages))

	actions := s.planActions(ctx, messages)
	var results []ChatSendResult

	logger.Info("Starting bulk chat message send")

	if dryRun {
		results = s.dryRunActions(actions)
	} else {
		results = s.executeActions(ctx, actions, ignoreError)
	}

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}

	logger.Info("Bulk chat message send completed",
		"successful", successCount,
		"failed", len(messages)-successCount)

	return results
}

func (s *chatSender) planActions(ctx context.Context, messages map[string]string) []*chatAction {
	actions := make([]*chatAction, 0, len(messages))

	for chatRef, content := range messages {
		processedContent, mentions, err := s.processMentions(ctx, chatRef, content)
		if err != nil {
			actions = append(actions, newChatErrorAction(chatRef, content, err))
			continue
		}

		messageData := chatMessageData{
			chatRef: chatRef,
			body: models.MessageBody{
				Content:     processedContent,
				ContentType: models.MessageContentTypeHTML,
				Mentions:    mentions,
			},
		}

		actions = append(actions, s.newSendAction(&messageData))
	}

	return actions
}

func (s *chatSender) processMentions(ctx context.Context, chatRef, content string) (string, []models.Mention, error) {
	rawMentions := templates.ExtractMentions(content)
	if len(rawMentions) == 0 {
		return content, nil, nil
	}

	ref := utils.GetChatRef(chatRef)
	mentions, err := s.chatService.GetMentions(ctx, ref, rawMentions)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve mentions for chat %s: %w", chatRef, err)
	}

	return templates.ReplaceMentions(content, mentions), mentions, nil
}

func newChatErrorAction(chatRef, content string, err error) *chatAction {
	return &chatAction{
		chatMessageData: &chatMessageData{
			chatRef: chatRef,
			body: models.MessageBody{
				Content: content,
			},
		},
		run:    nil,
		result: &ChatSendResult{ChatRef: chatRef, Message: content, Error: err},
	}
}

func (s *chatSender) newSendAction(data *chatMessageData) *chatAction {
	return &chatAction{
		chatMessageData: data,
		run:             s.sendMessage,
		result:          nil,
	}
}

func (s *chatSender) sendMessage(ctx context.Context, data *chatMessageData) *ChatSendResult {
	ref := utils.GetChatRef(data.chatRef)
	msg, err := s.chatService.SendMessage(ctx, ref, data.body)
	if err != nil {
		return &ChatSendResult{
			ChatRef: data.chatRef,
			Error:   fmt.Errorf("failed to send to chat %s: %w", data.chatRef, err),
		}
	}
	return &ChatSendResult{
		ChatRef: data.chatRef,
		Message: msg.Content,
	}
}

func (s *chatSender) executeActions(ctx context.Context, actions []*chatAction, ignoreError bool) []ChatSendResult {
	logger := initializers.Logger
	results := make([]ChatSendResult, 0, len(actions))
	skipRemaining := false

	for _, action := range actions {
		var result *ChatSendResult

		switch {
		case skipRemaining:
			logger.Debug("Skipping Message",
				"chat", action.chatRef)

			result = &ChatSendResult{
				ChatRef: action.chatRef,
				Message: action.body.Content,
				Error:   ErrMessageSkipped,
			}
		case action.run == nil:
			result = action.result
			logger.Error("Failed during planning",
				"chat", action.chatRef, "error", result.Error)
			if !ignoreError {
				skipRemaining = true
			}
		default:
			logger.Debug("Sending message to chat",
				"chat", action.chatRef)

			result = action.run(ctx, action.chatMessageData)

			if result.Error != nil {
				logger.Error("Failed to send message",
					"chat", action.chatRef, "error", result.Error)
				if !ignoreError {
					skipRemaining = true
				}
			} else {
				logger.Info("Message sent successfully", "chat", action.chatRef)
			}
		}

		results = append(results, *result)
	}

	return results
}

func (s *chatSender) dryRunActions(actions []*chatAction) []ChatSendResult {
	results := make([]ChatSendResult, 0, len(actions))
	for _, act := range actions {
		if act.run == nil {
			results = append(results, *act.result)
		} else {
			result := ChatSendResult{ChatRef: act.chatRef, Message: act.body.Content}
			results = append(results, result)
		}
	}

	return results
}
