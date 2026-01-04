package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/channels/sender"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var channelsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send messages to Teams channels",
	Long: `Send messages to one or more Teams channels using templates, raw strings, or text files.

Examples:
  # Send templated messages
  cli channels send --template msg.txt --data recipients.yaml --team MyTeam

  # Send raw message to specific channels
  cli channels send --message "Hello team!" --channels General,Announcements --team MyTeam

  # Send message from file
  cli channels send --message-file msg.txt --channels General --team MyTeam

  # Dry run to preview
  cli channels send --template msg.txt --data recipients.yaml --team MyTeam --dry-run`,
	RunE: runChannelsSend,
}

var (
	template     string
	templateData string
	message      string
	messageFile  string

	team     string
	channels []string

	channelsDryRun       bool
	channelsIgnoreErrors bool
)

func init() {
	channelsSendCmd.Flags().StringVar(&template, "template", "", "Path to message template file")
	channelsSendCmd.Flags().StringVar(&templateData, "data", "", "Path to data file (YAML/JSON/TOML/CSV)")
	channelsSendCmd.Flags().StringVar(&message, "message", "", "Raw message string")
	channelsSendCmd.Flags().StringVar(&messageFile, "message-file", "", "Path to text file containing message")

	channelsSendCmd.Flags().StringVar(&team, "team", "", "Team name or ID (required)")
	channelsSendCmd.Flags().StringSliceVar(&channels, "channels", nil, "Comma-separated list of channel names/IDs")

	channelsSendCmd.Flags().BoolVar(&channelsDryRun, "dry-run", false, "Preview messages without sending")
	channelsSendCmd.Flags().BoolVar(&channelsIgnoreErrors, "ignore-errors", false, "Continue on errors")

	if err := channelsSendCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runChannelsSend(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()

	messages, err := validateAndProcessTeamsFlags()
	if err != nil {
		return err
	}

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Sending messages to channels", "team", team, "count", len(messages), "dryRun", channelsDryRun)
	results := teamsClient.Channels.Send(ctx, team, messages, channelsDryRun, channelsIgnoreErrors)

	printChannelResults(results, channelsDryRun)

	return nil
}

func validateAndProcessTeamsFlags() (map[string]string, error) {
	inputMethods := 0
	if template != "" {
		inputMethods++
	}
	if message != "" {
		inputMethods++
	}
	if messageFile != "" {
		inputMethods++
	}

	if inputMethods == 0 {
		return nil, fmt.Errorf("must specify one of: --template, --message, or --message-file")
	}
	if inputMethods > 1 {
		return nil, fmt.Errorf("cannot use --template, --message, and --message-file together")
	}

	switch {
	case template != "":
		return processTemplate()
	case message != "":
		return processMessage(message)
	case messageFile != "":
		return processMessageFile()
	}

	return nil, fmt.Errorf("internal error: no input method selected")
}

func processTemplate() (map[string]string, error) {
	if templateData == "" {
		return nil, fmt.Errorf("--data is required when using --template")
	}
	return parseTemplateAndData(template, templateData)
}

func processMessage(message string) (map[string]string, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("--channels is required when using --message")
	}
	return createMessagesFromString(message, channels), nil
}

func processMessageFile() (map[string]string, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("--channels is required when using --message-file")
	}
	return createMessagesFromFile(messageFile, channels)
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
