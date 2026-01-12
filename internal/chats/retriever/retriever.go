package retriever

import (
	"context"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/cli/internal/utils"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/search"
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
func (r *retriever) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange, chatRef chats.ChatRef) ([]*ChatMessageWithContext, error) {
	messages, err := r.getMessagesInTimeRange(ctx, timeRange, chatRef)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *retriever) getMessagesInTimeRange(ctx context.Context, timeRange coreretriever.TimeRange, chatRef chats.ChatRef) ([]*ChatMessageWithContext, error) {
	var aggregatedSearchResults []*search.SearchResults
	var from int32 = 0
	var size int32 = 25
	var cRef chats.ChatRef
	if chatRef != nil {
		cRef = chatRef
	}
	searchConfig := &search.SearchConfig{
		MaxWorkers: 32,
	}
	for {
		searchOpts := &search.SearchMessagesOptions{
			StartTime: &timeRange.Start,
			EndTime:   &timeRange.End,
			NotFromMe: true,
			SearchPage: &search.SearchPage{
				From: &from,
				Size: &size,
			},
		}

		searchResult, err := r.chatService.SearchMessages(ctx, cRef, searchOpts, searchConfig)
		if err != nil {
			return nil, err
		}
		aggregatedSearchResults = append(aggregatedSearchResults, searchResult)
		if searchResult.NextFrom == nil {
			break
		}
		from = *searchResult.NextFrom
	}
	return r.processChatMessages(ctx, aggregatedSearchResults), nil
}

func (r *retriever) processChatMessages(ctx context.Context, searchResults []*search.SearchResults) []*ChatMessageWithContext {
	var results []*ChatMessageWithContext
	var nameByChatID = make(map[string]string)
	var chatTypeByChatID = make(map[string]string)
	for _, searchResult := range searchResults {
		for _, msg := range searchResult.Messages {
			if msg.ChatID == nil {
				continue
			}
			var chatName string
			var chatType string
			if name, exists := nameByChatID[*msg.ChatID]; exists {
				chatName = name
				chatType = chatTypeByChatID[*msg.ChatID]
			} else if msg.ChatID != nil {
				chatRef := utils.GetChatRef(*msg.ChatID)
				switch cr := chatRef.(type) {
				case chats.OneOnOneChatRef:
					chatName = msg.Message.From.DisplayName
					chatType = "OneOnOne"
				case chats.GroupChatRef:
					chat, err := r.chatService.GetChat(ctx, cr)
					if err != nil {
						continue
					}
					if chat.Topic != nil {
						chatName = *chat.Topic
					} else {
						chatName = *msg.ChatID
					}
					chatType = "Group"
				}
				nameByChatID[*msg.ChatID] = chatName
				chatTypeByChatID[*msg.ChatID] = chatType
			}
			results = append(results, &ChatMessageWithContext{
				ChatName: chatName,
				ChatID:   *msg.ChatID,
				ChatType: chatType,
				Message:  msg.Message,
			})
		}
	}
	return results
}
