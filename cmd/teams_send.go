package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/messaging"
)

var teamsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send messages to Teams channels",
	Long: `Send messages to one or more Teams channels using templates, raw strings, or text files.

Examples:
  # Send templated messages
  cli teams send --template msg.txt --data recipients.yaml --team MyTeam

  # Send raw message to specific channels
  cli teams send --message "Hello team!" --channels General,Announcements --team MyTeam

  # Send message from file
  cli teams send --message-file msg.txt --channels General --team MyTeam

  # Dry run to preview
  cli teams send --template msg.txt --data recipients.yaml --team MyTeam --dry-run`,
	RunE: runTeamsSend,
}

var (
	teamsTemplate    string
	teamsData        string
	teamsMessage     string
	teamsMessageFile string

	teamsTeam     string
	teamsChannels []string

	teamsDryRun       bool
	teamsIgnoreErrors bool
)

func init() {
	teamsSendCmd.Flags().StringVar(&teamsTemplate, "template", "", "Path to message template file")
	teamsSendCmd.Flags().StringVar(&teamsData, "data", "", "Path to data file (YAML/JSON/TOML/CSV)")
	teamsSendCmd.Flags().StringVar(&teamsMessage, "message", "", "Raw message string")
	teamsSendCmd.Flags().StringVar(&teamsMessageFile, "message-file", "", "Path to text file containing message")

	teamsSendCmd.Flags().StringVar(&teamsTeam, "team", "", "Team name or ID (required)")
	teamsSendCmd.Flags().StringSliceVar(&teamsChannels, "channels", nil, "Comma-separated list of channel names/IDs")

	teamsSendCmd.Flags().BoolVar(&teamsDryRun, "dry-run", false, "Preview messages without sending")
	teamsSendCmd.Flags().BoolVar(&teamsIgnoreErrors, "ignore-errors", false, "Continue on errors")

	if err := teamsSendCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runTeamsSend(cmd *cobra.Command, args []string) error {
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

	log.Info("Sending messages to channels", "team", teamsTeam, "count", len(messages), "dryRun", teamsDryRun)
	results := teamsClient.ChannelSender.SendToChannels(ctx, teamsTeam, messages, teamsDryRun, teamsIgnoreErrors)

	printChannelResults(results, teamsDryRun)

	return nil
}

func validateAndProcessTeamsFlags() (map[string]string, error) {
	inputMethods := 0
	if teamsTemplate != "" {
		inputMethods++
	}
	if teamsMessage != "" {
		inputMethods++
	}
	if teamsMessageFile != "" {
		inputMethods++
	}

	if inputMethods == 0 {
		return nil, fmt.Errorf("must specify one of: --template, --message, or --message-file")
	}
	if inputMethods > 1 {
		return nil, fmt.Errorf("cannot use --template, --message, and --message-file together")
	}

	switch {
	case teamsTemplate != "":
		return processTeamsTemplateMode()
	case teamsMessage != "":
		return processTeamsMessageMode(teamsMessage)
	case teamsMessageFile != "":
		return processTeamsMessageFileMode()
	}

	return nil, fmt.Errorf("internal error: no input method selected")
}

func processTeamsTemplateMode() (map[string]string, error) {
	if teamsData == "" {
		return nil, fmt.Errorf("--data is required when using --template")
	}
	return parseTemplateAndData(teamsTemplate, teamsData)
}

func processTeamsMessageMode(message string) (map[string]string, error) {
	if len(teamsChannels) == 0 {
		return nil, fmt.Errorf("--channels is required when using --message")
	}
	return createMessagesFromString(message, teamsChannels), nil
}

func processTeamsMessageFileMode() (map[string]string, error) {
	if len(teamsChannels) == 0 {
		return nil, fmt.Errorf("--channels is required when using --message-file")
	}
	return createMessagesFromFile(teamsMessageFile, teamsChannels)
}

func printChannelResults(results []messaging.ChannelSendResult, dryRun bool) {
	log := initializers.Logger
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				log.Warn("Would fail", "channel", res.ChannelRef, "error", res.Error)
			} else {
				log.Info("Would send", "channel", res.ChannelRef, "message", res.Message)
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.Error != nil {
				log.Error("Failed", "channel", res.ChannelRef, "error", res.Error)
			} else {
				successCount++
			}
		}
		log.Info("Send complete", "successful", successCount, "total", len(results))
	}
}
