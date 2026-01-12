package sender

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
	coresender "github.com/pzsp-teams/teams-cli/internal/core/sender"
)

type chatSender struct {
	generic *coresender.GenericSender[ChatSendResult]
}

// ChatSender defines the interface for sending messages to chats
type ChatSender interface {
	Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult
}

// NewChatSender creates a new ChatSender with the provided chats service
func NewChatSender(chatService chats.Service) ChatSender {
	adapter := &chatAdapter{
		chatService: chatService,
	}
	newResult := func(ref, message string, err error) ChatSendResult {
		return ChatSendResult{ChatRef: ref, Message: message, Error: err}
	}
	return &chatSender{
		generic: coresender.NewGenericSender(adapter, newResult, coresender.TypeChat),
	}
}

// Send sends messages to multiple chats
//
// messages: map of chat reference (name or ID) to message content
// Returns a slice of ChatSendResult containing the outcome for each chat
func (s *chatSender) Send(ctx context.Context, messages map[string]string, dryRun, ignoreError bool) []ChatSendResult {
	return s.generic.Send(ctx, messages, dryRun, ignoreError)
}
