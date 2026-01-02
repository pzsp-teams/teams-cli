package messaging

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

type senderType int

const (
	senderTypeChannel senderType = iota
	senderTypeChat
)

func (t senderType) String() string {
	switch t {
	case senderTypeChannel:
		return "channel"
	case senderTypeChat:
		return "chat"
	default:
		return "unknown"
	}
}

// sendResult is implemented by ChannelSendResult and ChatSendResult
type sendResult interface {
	getError() error
	getRef() string
	getMessage() string
}

// senderAdapter abstracts type-specific operations for channel and chat senders.
// Res: Result type
type senderAdapter[Res any] interface {
	getMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error)
	sendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error)
	newErrorResult(ref string, content string, err error) Res
	newSuccessResult(ref string, message string) Res
	getError(result Res) error
	getLogFields(ref string) map[string]any
}

type messageData struct {
	ref  string
	body models.MessageBody
}

type action[Res any] struct {
	data   *messageData
	run    func(ctx context.Context, data *messageData) Res
	result *Res
}
