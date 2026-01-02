package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/cli/internal/messaging"
)

var chatsSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send messages to Teams chats",
	Long: `Send messages to one or more Teams chats using templates, raw strings, or text files.

Examples:
  # Send templated messages
  cli chats send --template msg.txt --data recipients.yaml

  # Send raw message to specific chats
  cli chats send --message "Hello!" --chats user1@domain.com,user2@domain.com

  # Send message from file
  cli chats send --message-file msg.txt --chats user@domain.com

  # Dry run to preview
  cli chats send --template msg.txt --data recipients.yaml --dry-run`,
	RunE: runChatsSend,
}

var (
	chatsTemplate    string
	chatsData        string
	chatsMessage     string
	chatsMessageFile string

	chatsList []string

	chatsDryRun       bool
	chatsIgnoreErrors bool
)

func init() {
	chatsSendCmd.Flags().StringVar(&chatsTemplate, "template", "", "Path to message template file")
	chatsSendCmd.Flags().StringVar(&chatsData, "data", "", "Path to data file (YAML/JSON/TOML/CSV)")
	chatsSendCmd.Flags().StringVar(&chatsMessage, "message", "", "Raw message string")
	chatsSendCmd.Flags().StringVar(&chatsMessageFile, "message-file", "", "Path to text file containing message")

	chatsSendCmd.Flags().StringSliceVar(&chatsList, "chats", nil, "Comma-separated list of chat recipients (email/ID)")

	chatsSendCmd.Flags().BoolVar(&chatsDryRun, "dry-run", false, "Preview messages without sending")
	chatsSendCmd.Flags().BoolVar(&chatsIgnoreErrors, "ignore-errors", false, "Continue on errors")
}

func runChatsSend(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := context.TODO()

	messages, err := validateAndProcessChatsFlags()
	if err != nil {
		return err
	}

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Sending messages to chats", "count", len(messages), "dryRun", chatsDryRun)
	results := teamsClient.ChatSender.Send(ctx, messages, chatsDryRun, chatsIgnoreErrors)

	printChatResults(results, chatsDryRun)

	return nil
}

func validateAndProcessChatsFlags() (map[string]string, error) {
	inputMethods := 0
	if chatsTemplate != "" {
		inputMethods++
	}
	if chatsMessage != "" {
		inputMethods++
	}
	if chatsMessageFile != "" {
		inputMethods++
	}

	if inputMethods == 0 {
		return nil, fmt.Errorf("must specify one of: --template, --message, or --message-file")
	}
	if inputMethods > 1 {
		return nil, fmt.Errorf("cannot use --template, --message, and --message-file together")
	}

	switch {
	case chatsTemplate != "":
		return processChatsTemplateMode()
	case chatsMessage != "":
		return processChatsMessageMode(chatsMessage)
	case chatsMessageFile != "":
		return processChatsMessageFileMode()
	}

	return nil, fmt.Errorf("internal error: no input method selected")
}

func processChatsTemplateMode() (map[string]string, error) {
	if chatsData == "" {
		return nil, fmt.Errorf("--data is required when using --template")
	}
	return parseTemplateAndData(chatsTemplate, chatsData)
}

func processChatsMessageMode(message string) (map[string]string, error) {
	if len(chatsList) == 0 {
		return nil, fmt.Errorf("--chats is required when using --message")
	}
	return createMessagesFromString(message, chatsList), nil
}

func processChatsMessageFileMode() (map[string]string, error) {
	if len(chatsList) == 0 {
		return nil, fmt.Errorf("--chats is required when using --message-file")
	}
	return createMessagesFromFile(chatsMessageFile, chatsList)
}

func printChatResults(results []messaging.ChatSendResult, dryRun bool) {
	if dryRun {
		for _, res := range results {
			if res.Error != nil {
				fmt.Printf("Would fail - chat: %s, error: %v\n", res.ChatRef, res.Error)
			} else {
				fmt.Printf("Would send - chat: %s, message: %s\n", res.ChatRef, res.Message)
			}
		}
	} else {
		successCount := 0
		for _, res := range results {
			if res.Error != nil {
				fmt.Printf("Failed - chat: %s, error: %v\n", res.ChatRef, res.Error)
			} else {
				successCount++
			}
		}
		fmt.Printf("Send complete - successful: %d, total: %d\n", successCount, len(results))
	}
}
