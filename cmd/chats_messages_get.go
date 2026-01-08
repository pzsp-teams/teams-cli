package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	chatsretriever "github.com/pzsp-teams/cli/internal/chats/retriever"
	"github.com/pzsp-teams/cli/internal/formatters"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var chatsMessagesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve messages from all chats within a time range",
	Long: `Retrieve messages from all chats you have access to within the specified time range.

Examples:
  # Last 24 hours (default)
  teams-cli chats messages get

  # From 2 hours ago till now
  teams-cli chats messages get --start "2 hours ago"

  # Specific time window
  teams-cli chats messages get --start "2024-01-01 10:00" --end "2024-01-01 11:00"

  # Yesterday
  teams-cli chats messages get --start yesterday --end now`,
	RunE: runChatsMessagesGet,
}

var (
	chatsMessagesStartTime string
	chatsMessagesEndTime   string
	chatMessagesFile       string
	chatMessagesFormat     formatFlags
)

func init() {
	chatsMessagesGetCmd.Flags().StringVar(&chatsMessagesStartTime, "start", "", "Start time (optional, defaults to 24h ago)")
	chatsMessagesGetCmd.Flags().StringVar(&chatsMessagesEndTime, "end", "", "End time (optional, defaults to now)")
	chatsMessagesGetCmd.Flags().StringVar(&chatMessagesFile, "file", "", "Name of file to save messages, will print to stdout if not given or apend to existing file if given")
	chatsMessagesGetCmd.Flags().BoolVar(&chatMessagesFormat.Plain, "plain", false, "Use plain format")
	chatsMessagesGetCmd.Flags().BoolVar(&chatMessagesFormat.MarkDown, "markdown", false, "Use Markdown format")
	chatsMessagesGetCmd.MarkFlagsMutuallyExclusive("plain", "markdown")
}

func runChatsMessagesGet(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()
	dest := getDest(chatMessagesFile)
	defer func() {
		if err := dest.Close(); err != nil {
			log.Error("Failed to close destination", "error", err)
		}
	}()

	formatter, err := chatMessagesFormat.getFormatter()
	if err != nil {
		return err
	}

	timeRange, err := ParseTimeRange(chatsMessagesStartTime, chatsMessagesEndTime)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Retrieving chat messages", "start", timeRange.Start, "end", timeRange.End)
	messages, err := teamsClient.Chats.GetMessages(ctx, timeRange)
	if err != nil {
		log.Error("Failed to retrieve messages", "error", err)
		return err
	}

	log.Info("Retrieved messages", "count", len(messages))

	viewMessages := make([]*formatters.MessageView, 0, len(messages))
	for _, m := range messages {
		viewMessages = append(viewMessages, toChatMessageView(m))
	}

	if err := formatter.WriteMessages(dest, viewMessages); err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
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