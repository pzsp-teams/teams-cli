package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/initializers"
)

var channelsMessagesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve messages from all channels within a time range",
	Long: `Retrieve messages from all channels you have access to within the specified time range.

Examples:
  # Last 24 hours (default)
  teams-cli channels messages get

  # From 2 hours ago till now
  teams-cli channels messages get --start "2 hours ago"

  # Specific time window
  teams-cli channels messages get --start "2024-01-01 10:00" --end "2024-01-01 11:00"

  # Yesterday
  teams-cli channels messages get --start yesterday --end now`,
	RunE: runChannelsMessagesGet,
}

var (
	channelsMessagesStartTime string
	channelsMessagesEndTime   string
)

func init() {
	channelsMessagesGetCmd.Flags().StringVar(&channelsMessagesStartTime, "start", "", "Start time (optional, defaults to 24h ago)")
	channelsMessagesGetCmd.Flags().StringVar(&channelsMessagesEndTime, "end", "", "End time (optional, defaults to now)")
}

func runChannelsMessagesGet(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()

	timeRange, err := ParseTimeRange(channelsMessagesStartTime, channelsMessagesEndTime)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Retrieving channel messages", "start", timeRange.Start, "end", timeRange.End)
	messages, err := teamsClient.Channels.GetMessages(ctx, timeRange)
	if err != nil {
		log.Error("Failed to retrieve messages", "error", err)
		return err
	}

	log.Info("Retrieved messages", "count", len(messages))
	printChannelMessages(messages)

	return nil
}
