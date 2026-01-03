package sender

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

// Type represents the type of sender (channel or chat)
type Type int

const (
	// TypeChannel represents a channel sender
	TypeChannel Type = iota
	// TypeChat represents a chat sender
	TypeChat
)

// String returns the string representation of the sender type
func (t Type) String() string {
	switch t {
	case TypeChannel:
		return "channel"
	case TypeChat:
		return "chat"
	default:
		return "unknown"
	}
}

// Result is implemented by ChannelSendResult and ChatSendResult
type Result interface {
	GetError() error
	GetRef() string
	GetMessage() string
}

// Adapter abstracts type-specific operations for channel and chat senders.
type Adapter interface {
	GetMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error)
	SendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error)
}

// MessageData holds the reference and body for a message to be sent
type MessageData struct {
	ref  string
	body models.MessageBody
}

// Action represents a planned message send operation
type Action[Res any] struct {
	data   *MessageData
	run    func(ctx context.Context, data *MessageData) Res
	result *Res
}
