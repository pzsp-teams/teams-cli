package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

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
)

func init() {
	chatsMessagesGetCmd.Flags().StringVar(&chatsMessagesStartTime, "start", "", "Start time (optional, defaults to 24h ago)")
	chatsMessagesGetCmd.Flags().StringVar(&chatsMessagesEndTime, "end", "", "End time (optional, defaults to now)")
}

func runChatsMessagesGet(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()

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
	printChatMessages(messages)

	return nil
}
