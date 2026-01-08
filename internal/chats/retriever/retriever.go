package retriever

import (
	"context"
	"fmt"
	"strings"
	"time"

	coreretriever "github.com/pzsp-teams/cli/internal/core/retriever"
	"github.com/pzsp-teams/lib/chats"
	"github.com/pzsp-teams/lib/models"
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

// getChatRef builds a ChatRef from a Chat model
func getChatRef(chat *models.Chat) chats.ChatRef {
	if chat.Type == models.ChatTypeOneOnOne {
		return chats.OneOnOneChatRef{Ref: chat.ID}
	}
	return chats.GroupChatRef{Ref: chat.ID}
}

type chatJob struct {
	Chat *models.Chat
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

	jobs := make([]chatJob, len(chatList))
	for i, chat := range chatList {
		jobs[i] = chatJob{Chat: chat}
	}

	results := coreretriever.ExecuteJobs(jobs, coreretriever.WorkersCount, func(job chatJob) ([]*ChatMessageWithContext, error) {
		return r.processChatMessages(ctx, job, timeRange)
	})

	return coreretriever.AggregateResults(results)
}

func (r *retriever) processChatMessages(ctx context.Context, job chatJob, timeRange coreretriever.TimeRange) ([]*ChatMessageWithContext, error) {
	chatRef := getChatRef(job.Chat)

	var messageCollection *models.MessageCollection
	var err error
	var filteredMessages []*ChatMessageWithContext
	var oldestMessage time.Time
	var nextLink *string

	for {
		messageCollection, err = r.chatService.ListMessages(ctx, chatRef, false, nextLink)
		if err != nil {
			if strings.Contains(err.Error(), "403") {
				return nil, nil // Skip forbidden chats
			}
			return nil, fmt.Errorf("%w: chat=%s: %v", ErrListingMessagesFailed, job.Chat.ID, err)
		}

		if len(messageCollection.Messages) == 0 {
			break
		}

		for _, msg := range messageCollection.Messages {
			if msg.CreatedDateTime.After(timeRange.Start) && msg.CreatedDateTime.Before(timeRange.End) {
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

		oldestMessage = messageCollection.Messages[len(messageCollection.Messages)-1].CreatedDateTime
		if oldestMessage.Before(timeRange.Start) {
			break
		}

		nextLink = messageCollection.NextLink
		if nextLink == nil || *nextLink == "" {
			break
		}
	}

	return filteredMessages, nil
}
