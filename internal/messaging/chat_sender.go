package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
)

type chatSender struct {
	generic *genericSender[ChatSendResult]
}

// ChatSender defines the interface for sending messages to chats
type ChatSender interface {
	Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult
}

// NewChatSender creates a new ChatSender with the provided chats service
func NewChatSender(chatService chats.Service) *chatSender {
	adapter := &chatAdapter{
		chatService: chatService,
	}
	return &chatSender{
		generic: newGenericSender(adapter),
	}
}

// Send sends messages to multiple chats
//
// messages: map of chat reference (name or ID) to message content
// Returns a slice of ChatSendResult containing the outcome for each chat
func (s *chatSender) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult {
	return s.generic.send(ctx, messages, dryRun, ignoreError)
}
