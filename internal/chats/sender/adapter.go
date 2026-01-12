package sender

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/teams-cli/internal/utils"
)

type chatAdapter struct {
	chatService chats.Service
}

// GetMentions implements the coresender.Adapter interface
func (a *chatAdapter) GetMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
	chatRef := utils.GetChatRef(ref)
	return a.chatService.GetMentions(ctx, chatRef, rawMentions)
}

// SendMessage implements the coresender.Adapter interface
func (a *chatAdapter) SendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
	chatRef := utils.GetChatRef(ref)
	return a.chatService.SendMessage(ctx, chatRef, body)
}
