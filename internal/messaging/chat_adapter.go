package messaging

import (
	"context"

	"github.com/pzsp-teams/cli/internal/utils"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

type chatAdapter struct {
	chatService chats.Service
}

func (a *chatAdapter) getMentions(ctx context.Context, ref string, rawMentions []string) ([]models.Mention, error) {
	chatRef := utils.GetChatRef(ref)
	return a.chatService.GetMentions(ctx, chatRef, rawMentions)
}

func (a *chatAdapter) sendMessage(ctx context.Context, ref string, body models.MessageBody) (*models.Message, error) {
	chatRef := utils.GetChatRef(ref)
	return a.chatService.SendMessage(ctx, chatRef, body)
}
