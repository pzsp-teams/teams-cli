package messages

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/cmd/common"
	chatsretriever "github.com/pzsp-teams/cli/internal/chats/retriever"
	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/cli/internal/formatters"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/utils"
)

// GetMessages handles retrieving messages from chats.
func GetMessages(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	startTime, _ := flags["start"].(string)
	endTime, _ := flags["end"].(string)
	file, _ := flags["file"].(string)
	format, _ := flags["format"].(string)
	chat, _ := flags["chat-ref"].(string)

	log := initializers.Logger

	var dest io.WriteCloser
	var err error
	if file != "" {
		dest, err = common.GetDest(file)
		if err != nil {
			return nil, err
		}
	} else {
		dest = common.NopWriteCloser(w)
	}

	defer func() {
		if err := dest.Close(); err != nil {
			log.Error("Failed to close destination", "error", err)
		}
	}()

	formatter := common.GetFormatter(format)
	if err != nil {
		return nil, err
	}

	timeRange, err := common.ParseTimeRange(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time range: %w", err)
	}

	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	log.Info("Retrieving chat messages", "start", timeRange.Start, "end", timeRange.End)
	messages, err := c.Chats.GetMessages(ctx, timeRange, utils.GetChatRef(chat))
	if err != nil {
		log.Error("Failed to retrieve messages", "error", err)
		return nil, err
	}

	log.Info("Retrieved messages", "count", len(messages))

	viewMessages := make([]*formatters.MessageView, 0, len(messages))
	for _, m := range messages {
		viewMessages = append(viewMessages, toChatMessageView(m))
	}

	if err := formatter.WriteMessages(dest, viewMessages); err != nil {
		return nil, fmt.Errorf("failed to write messages: %w", err)
	}

	return messages, nil
}

func toChatMessageView(m *chatsretriever.ChatMessageWithContext) *formatters.MessageView {
	return &formatters.MessageView{
		ID:        m.Message.ID,
		Author:    m.Message.From.DisplayName,
		Timestamp: m.Message.CreatedDateTime,
		Content:   m.Message.Content,
		Context: []formatters.ContextItem{
			{Label: "Chat", Value: m.ChatName},
			{Label: "Type", Value: m.ChatType},
		},
	}
}
