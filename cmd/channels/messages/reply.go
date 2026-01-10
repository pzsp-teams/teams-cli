package messages

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdcommon "github.com/pzsp-teams/cli/cmd/common"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

type replyFlags struct {
	team      string
	channel   string
	messageID string
	message   string
	file      string
}

func newReplyCommand() *cobra.Command {
	flags := &replyFlags{}

	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Send a reply to a message in a Teams channel",
		Long: `Send a reply to a specific message in a Teams channel using raw strings or text files.

Examples:
  # Send a reply with raw message
  cli channels messages reply --team MyTeam --channel General --message-id 123456 --message "Thanks for the update!"

  # Send a reply from file
  cli channels messages reply --team MyTeam --channel General --message-id 123456 --message-file reply.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReply(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "Team name or ID (required)")
	cmd.Flags().StringVar(&flags.channel, "channel", "", "Channel name or ID (required)")
	cmd.Flags().StringVar(&flags.messageID, "message-id", "", "ID of the message to reply to (required)")
	cmd.Flags().StringVar(&flags.message, "message", "", "Reply message string")
	cmd.Flags().StringVar(&flags.file, "message-file", "", "Path to text file containing reply message")

	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("channel"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark channel flag as required: %v\n", err)
	}
	if err := cmd.MarkFlagRequired("message-id"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark message-id flag as required: %v\n", err)
	}

	cmd.MarkFlagsMutuallyExclusive("message", "message-file")

	return cmd
}

func runReply(cmd *cobra.Command, flags *replyFlags) error {
	log := initializers.Logger
	ctx := cmd.Context()

	content, err := internalcommon.GetMessageContent(flags.message, flags.file)
	if err != nil {
		return err
	}

	teamsClient, err := cmdcommon.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	log.Info("Sending reply to message", "team", flags.team, "channel", flags.channel, "messageID", flags.messageID)
	result := teamsClient.Channels.SendReply(ctx, flags.team, flags.channel, flags.messageID, content)

	if result.Error != nil {
		fmt.Printf("Failed to send reply: %v\n", result.Error)
		return result.Error
	}

	fmt.Printf("Reply sent successfully to message %s in channel %s\n", flags.messageID, flags.channel)
	return nil
}
