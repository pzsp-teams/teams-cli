package retriever

import (
	"context"
	"fmt"
	"strings"

	f "github.com/pzsp-teams/cli/internal/core/formatter"
	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
)

type retriever struct {
	chatService chats.Service
	formatter   f.Formatter
}

// NewRetriever creates a new chat message retriever
func NewRetriever(chatService chats.Service, formatter f.Formatter) Retriever {
	return &retriever{
		chatService: chatService,
		formatter:   formatter,
	}
}

// getChatRef builds a ChatRef from a Chat model
func getChatRef(chat *models.Chat) chats.ChatRef {
	if chat.Type == models.ChatTypeOneOnOne {
		return chats.OneOnOneChatRef{Ref: chat.ID}
	}
	return chats.GroupChatRef{Ref: chat.ID}
}

// GetMessages retrieves messages from all chats within the specified time range
func (r *retriever) GetMessages(ctx context.Context, timeRange coreretriever.TimeRange) ([]*ChatMessageWithContext, error) {
	chatList, err := r.chatService.ListChats(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrListingChatsFailed, err)
	}

	if len(chatList) == 0 {
		return nil, ErrNoChatsFound
	}

	type chatJob struct {
		Chat *models.Chat
	}

	jobs := make([]chatJob, len(chatList))
	for i, chat := range chatList {
		jobs[i] = chatJob{Chat: chat}
	}

	results := coreretriever.ExecuteJobs(jobs, coreretriever.WorkersCount, func(job chatJob) ([]*ChatMessageWithContext, error) {
		chatRef := getChatRef(job.Chat)
		messages, err := r.chatService.ListMessages(ctx, chatRef, false)

		if err != nil && !strings.Contains(err.Error(), "403") {
			return nil, fmt.Errorf("%w: chat=%s: %v", ErrListingMessagesFailed, job.Chat.ID, err)
		}

		var filteredMessages []*ChatMessageWithContext
		for _, msg := range messages {
			if msg.CreatedDateTime.After(timeRange.Start) && msg.CreatedDateTime.Before(timeRange.End) {
				if msg.ContentType == models.MessageContentTypeHTML {
					msg.Content = r.formatter.Format(msg.Content)
				}

				chatName := job.Chat.ID
				if job.Chat.Topic != nil && *job.Chat.Topic != "" {
					chatName = *job.Chat.Topic
				}

				filteredMessages = append(filteredMessages, &ChatMessageWithContext{
					ChatName: chatName,
					ChatID:   job.Chat.ID,
					ChatType: string(job.Chat.Type),
					Message:  msg,
				})
			}
		}

		return filteredMessages, nil
	})

	return coreretriever.AggregateResults(results)
}
