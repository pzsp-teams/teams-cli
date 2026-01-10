package messages

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdcommon "github.com/pzsp-teams/cli/cmd/common"
	channelsretriever "github.com/pzsp-teams/cli/internal/channels/retriever"
	"github.com/pzsp-teams/cli/internal/formatters"
	"github.com/pzsp-teams/cli/internal/initializers"
)

type getFlags struct {
	startTime        string
	endTime          string
	file             string
	plain            bool
	markdown         bool
	teamReference    string
	channelReference string
}

func newGetCommand() *cobra.Command {
	flags := &getFlags{}

	cmd := &cobra.Command{
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
  teams-cli channels messages get --start yesterday --end now
	
  # Filter by team
  teams-cli channels messages get --team "My Team"

  # Filter by channel and team
  teams-cli channels messages get --team "My Team" --channel "General"
`,

		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.startTime, "start", "", "Start time (optional, defaults to 24h ago)")
	cmd.Flags().StringVar(&flags.endTime, "end", "", "End time (optional, defaults to now)")
	cmd.Flags().StringVar(&flags.file, "file", "", "Name of file to save messages, will print to stdout if not given or apend to existing file if given")
	cmd.Flags().BoolVar(&flags.plain, "plain", false, "Use plain format")
	cmd.Flags().BoolVar(&flags.markdown, "markdown", false, "Use Markdown format")
	cmd.Flags().StringVar(&flags.teamReference, "team", "", "Team ID or name to filter messages (optional)")
	cmd.Flags().StringVar(&flags.channelReference, "channel", "", "Channel ID or name to filter messages (optional)")
	cmd.MarkFlagsMutuallyExclusive("plain", "markdown")

	return cmd
}

func runGet(cmd *cobra.Command, flags *getFlags) error {
	log := initializers.Logger
	ctx := cmd.Context()
	dest := cmdcommon.GetDest(flags.file)
	defer func() {
		if err := dest.Close(); err != nil {
			log.Error("Failed to close destination", "error", err)
		}
	}()

	formatter, err := cmdcommon.GetFormatter(flags.plain, flags.markdown)
	if err != nil {
		return err
	}

	timeRange, err := cmdcommon.ParseTimeRange(flags.startTime, flags.endTime)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	teamsClient, err := cmdcommon.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	log.Info("Retrieving channel messages", "start", timeRange.Start, "end", timeRange.End)
	messages, err := teamsClient.Channels.GetMessages(ctx, timeRange, &flags.teamReference, &flags.channelReference)
	if err != nil {
		log.Error("Failed to retrieve messages", "error", err)
		return err
	}

	log.Info("Retrieved messages", "count", len(messages))

	viewMessages := make([]*formatters.MessageView, 0, len(messages))
	for _, m := range messages {
		viewMessages = append(viewMessages, toChannelMessageView(m))
	}

	if err := formatter.WriteMessages(dest, viewMessages); err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	return nil
}

func toChannelMessageView(m *channelsretriever.ChannelMessageWithContext) *formatters.MessageView {
	return &formatters.MessageView{
		ID:        m.Message.ID,
		Author:    m.Message.From.DisplayName,
		Timestamp: m.Message.CreatedDateTime,
		Content:   m.Message.Content,
		Context: []formatters.ContextItem{
			{Label: "Team", Value: m.TeamName},
			{Label: "Channel", Value: m.ChannelName},
		},
	}
}
