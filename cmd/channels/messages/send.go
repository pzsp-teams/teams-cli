package messages

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdcommon "github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/cli/internal/channels/sender"
	internalcommon "github.com/pzsp-teams/cli/internal/common"
	"github.com/pzsp-teams/cli/internal/initializers"
)

type sendFlags struct {
	template     string
	templateData string
	message      string
	messageFile  string

	team     string
	channels []string

	dryRun       bool
	ignoreErrors bool
}

func newSendCommand() *cobra.Command {
	flags := &sendFlags{}

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages to Teams channels",
		Long: `Send messages to one or more Teams channels using templates, raw strings, or text files.

Examples:
  # Send templated messages
  cli channels messages send --template msg.txt --data recipients.yaml --team MyTeam

  # Send raw message to specific channels
  cli channels messages send --message "Hello team!" --channels General,Announcements --team MyTeam

  # Send message from file
  cli channels messages send --message-file msg.txt --channels General --team MyTeam

  # Dry run to preview
  cli channels messages send --template msg.txt --data recipients.yaml --team MyTeam --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSend(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.template, "template", "", "Path to message template file")
	cmd.Flags().StringVar(&flags.templateData, "data", "", "Path to data file (YAML/JSON/TOML/CSV)")
	cmd.Flags().StringVar(&flags.message, "message", "", "Raw message string")
	cmd.Flags().StringVar(&flags.messageFile, "message-file", "", "Path to text file containing message")

	cmd.Flags().StringVar(&flags.team, "team", "", "Team name or ID (required)")
	cmd.Flags().StringSliceVar(&flags.channels, "channels", nil, "Comma-separated list of channel names/IDs")

	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview messages without sending")
	cmd.Flags().BoolVar(&flags.ignoreErrors, "ignore-errors", false, "Continue on errors")

	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runSend(cmd *cobra.Command, flags *sendFlags) error {
	log := initializers.Logger
	ctx := cmd.Context()

	inputFlags := cmdcommon.MessageInputFlags{
		Template:     flags.template,
		TemplateData: flags.templateData,
		Message:      flags.message,
		MessageFile:  flags.messageFile,
	}

	processed, err := cmdcommon.ProcessMessageFlags(
		inputFlags,
		flags.channels,
		internalcommon.ParseTemplateAndData, // Dependency injection for testability
	)
	if err != nil {
		return err
	}

	teamsClient, err := cmdcommon.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	log.Info("Sending messages to channels", "team", flags.team, "count", len(processed.Messages), "dryRun", flags.dryRun)
	results := teamsClient.Channels.Send(ctx, flags.team, processed.Messages, flags.dryRun, flags.ignoreErrors)

	printChannelResults(results, flags.dryRun)

	return nil
}

func printChannelResults(results []sender.ChannelSendResult, dryRun bool) {
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				fmt.Printf("Would fail - channel: %s, error: %v\n", res.ChannelRef, res.Error)
			} else {
				fmt.Printf("Would send - channel: %s, message: %s\n", res.ChannelRef, res.Message)
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.Error != nil {
				fmt.Printf("Failed - channel: %s, error: %v\n", res.ChannelRef, res.Error)
			} else {
				successCount++
			}
		}
		fmt.Printf("Send complete - successful: %d, total: %d\n", successCount, len(results))
	}
}
