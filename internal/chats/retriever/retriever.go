package retriever

import (
	"context"

	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/search"
	coreretriever "github.com/pzsp-teams/teams-cli/internal/core/retriever"
	"github.com/pzsp-teams/teams-cli/internal/utils"
)

type retriever struct {
	chatService chats.Service
}

// NewRetriever creates a new chat message retriever
func NewRetriever(chatService chats.Service) Retriever {
	return &retriever{
		chatService: chatService,
	}
}

// GetMessages retrieves messages from all chats within the specified time range
func (r *retriever) GetMessages(
	ctx context.Context,
	timeRange coreretriever.TimeRange,
	chatRef chats.ChatRef,
) ([]*ChatMessageWithContext, error) {
	searchConfig := &search.SearchConfig{
		MaxWorkers: coreretriever.WorkersCount,
	}

	nameByChatID := make(map[string]string)
	typeByChatID := make(map[string]string)

	var out []*ChatMessageWithContext

	from := int32(0)
	size := int32(25)

	for range 10_000 {
		f := from
		s := size

		searchOpts := &search.SearchMessagesOptions{
			StartTime: &timeRange.Start,
			EndTime:   &timeRange.End,
			NotFromMe: true,
			SearchPage: &search.SearchPage{
				From: &f,
				Size: &s,
			},
			ToMe: true,
		}

		page, err := r.chatService.SearchMessages(ctx, chatRef, searchOpts, searchConfig)
		if err != nil {
			return nil, err
		}

		r.processPage(ctx, page, nameByChatID, typeByChatID, &out)

		if page.NextFrom == nil || *page.NextFrom == from {
			break
		}
		from = *page.NextFrom
	}

	return out, nil
}

func (r *retriever) processPage(
	ctx context.Context,
	page *search.SearchResults,
	nameByChatID map[string]string,
	typeByChatID map[string]string,
	out *[]*ChatMessageWithContext,
) {
	for _, msg := range page.Messages {
		if msg.ChatID == nil {
			continue
		}
		chatID := *msg.ChatID

		chatName := nameByChatID[chatID]
		chatType := typeByChatID[chatID]

		if chatName == "" {
			chatName, chatType = r.resolveChatMeta(ctx, chatID, msg)
			nameByChatID[chatID] = chatName
			typeByChatID[chatID] = chatType
		}

		*out = append(*out, &ChatMessageWithContext{
			ChatName: chatName,
			ChatID:   chatID,
			ChatType: chatType,
			Message:  msg.Message,
		})
	}
}

func (r *retriever) resolveChatMeta(ctx context.Context, chatID string, msg *search.SearchResult) (chatName, chatType string) {
	ref := utils.GetChatRef(chatID)

	switch cr := ref.(type) {
	case chats.OneOnOneChatRef:
		if msg.Message != nil && msg.Message.From != nil && msg.Message.From.DisplayName != "" {
			return msg.Message.From.DisplayName, "OneOnOne"
		}
		return chatID, "OneOnOne"

	case chats.GroupChatRef:
		chat, err := r.chatService.GetChat(ctx, cr)
		if err != nil || chat == nil {
			return chatID, "Group"
		}
		if chat.Topic != nil && *chat.Topic != "" {
			return *chat.Topic, "Group"
		}
		return chatID, "Group"
	}

	return chatID, "Unknown"
}
