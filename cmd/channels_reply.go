package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var channelsReplyCmd = &cobra.Command{
	Use:   "reply",
	Short: "Send a reply to a message in a Teams channel",
	Long: `Send a reply to a specific message in a Teams channel using raw strings or text files.

Examples:
  # Send a reply with raw message
  cli channels reply --team MyTeam --channel General --message-id 123456 --message "Thanks for the update!"

  # Send a reply from file
  cli channels reply --team MyTeam --channel General --message-id 123456 --message-file reply.txt`,
	RunE: runChannelsReply,
}

var (
	replyTeam      string
	replyChannel   string
	replyMessageID string
	replyMessage   string
	replyFile      string
)

func init() {
	channelsReplyCmd.Flags().StringVar(&replyTeam, "team", "", "Team name or ID (required)")
	channelsReplyCmd.Flags().StringVar(&replyChannel, "channel", "", "Channel name or ID (required)")
	channelsReplyCmd.Flags().StringVar(&replyMessageID, "message-id", "", "ID of the message to reply to (required)")
	channelsReplyCmd.Flags().StringVar(&replyMessage, "message", "", "Reply message string")
	channelsReplyCmd.Flags().StringVar(&replyFile, "message-file", "", "Path to text file containing reply message")

	if err := channelsReplyCmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}
	if err := channelsReplyCmd.MarkFlagRequired("channel"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark channel flag as required: %v\n", err)
	}
	if err := channelsReplyCmd.MarkFlagRequired("message-id"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark message-id flag as required: %v\n", err)
	}

	channelsReplyCmd.MarkFlagsMutuallyExclusive("message", "message-file")
}

func runChannelsReply(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := cmd.Context()

	content, err := getMessageContent(replyMessage, replyFile)
	if err != nil {
		return err
	}

	log.Debug("Creating Teams client")
	teamsClient, err := common.GetTeamsClient(cmd)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Sending reply to message", "team", replyTeam, "channel", replyChannel, "messageID", replyMessageID)
	result := teamsClient.Channels.SendReply(ctx, replyTeam, replyChannel, replyMessageID, content)

	if result.Error != nil {
		fmt.Printf("Failed to send reply: %v\n", result.Error)
		return result.Error
	}

	fmt.Printf("Reply sent successfully to message %s in channel %s\n", replyMessageID, replyChannel)
	return nil
}
