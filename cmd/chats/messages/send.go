package messages

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdcommon "github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/cli/internal/chats/sender"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

type sendFlags struct {
	template    string
	data        string
	message     string
	messageFile string

	chats []string

	dryRun       bool
	ignoreErrors bool
}

func newSendCommand() *cobra.Command {
	flags := &sendFlags{}

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages to Teams chats",
		Long: `Send messages to one or more Teams chats using templates, raw strings, or text files.

Examples:
  # Send templated messages
  cli chats messages send --template msg.txt --data recipients.yaml

  # Send raw message to specific chats
  cli chats messages send --message "Hello!" --chats user1@domain.com,user2@domain.com

  # Send message from file
  cli chats messages send --message-file msg.txt --chats user@domain.com

  # Dry run to preview
  cli chats messages send --template msg.txt --data recipients.yaml --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.template, "template", "", "Path to message template file")
	cmd.Flags().StringVar(&flags.data, "data", "", "Path to data file (YAML/JSON/TOML/CSV)")
	cmd.Flags().StringVar(&flags.message, "message", "", "Raw message string")
	cmd.Flags().StringVar(&flags.messageFile, "message-file", "", "Path to text file containing message")

	cmd.Flags().StringSliceVar(&flags.chats, "chats", nil, "Comma-separated list of chat recipients (email/ID)")

	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview messages without sending")
	cmd.Flags().BoolVar(&flags.ignoreErrors, "ignore-errors", false, "Continue on errors")

	return cmd
}

func runSend(cmd *cobra.Command, flags *sendFlags) error {
	log := initializers.Logger
	ctx := cmd.Context()

	inputFlags := cmdcommon.MessageInputFlags{
		Template:     flags.template,
		TemplateData: flags.data,
		Message:      flags.message,
		MessageFile:  flags.messageFile,
	}

	processed, err := cmdcommon.ProcessMessageFlags(
		inputFlags,
		flags.chats,
		internalcommon.ParseTemplateAndData,
	)
	if err != nil {
		return err
	}

	teamsClient, err := cmdcommon.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	log.Info("Sending messages to chats", "count", len(processed.Messages), "dryRun", flags.dryRun)
	results := teamsClient.Chats.Send(ctx, processed.Messages, flags.dryRun, flags.ignoreErrors)

	printChatResults(results, flags.dryRun)

	return nil
}

func printChatResults(results []sender.ChatSendResult, dryRun bool) {
	if dryRun {
		for _, res := range results {
			if res.GetError() != nil {
				fmt.Printf("Would fail - chat: %s, error: %v\n", res.GetRef(), res.GetError())
			} else {
				fmt.Printf("Would send - chat: %s, message: %s\n", res.GetRef(), res.GetMessage())
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.GetError() != nil {
				fmt.Printf("Failed - chat: %s, error: %v\n", res.GetRef(), res.GetError())
			} else {
				successCount++
			}
		}
		fmt.Printf("Send complete - successful: %d, total: %d\n", successCount, len(results))
	}
}
