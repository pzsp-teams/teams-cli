package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/internal/initializers"
)

var chatsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all chats",
	Long: `List all chats you have access to.

Examples:
  # List all chats
  teams-cli chats list`,
	RunE: runChatsList,
}

func runChatsList(cmd *cobra.Command, args []string) error {
	log := initializers.Logger
	ctx := cmd.Context()

	log.Debug("Creating Teams client")
	teamsClient, err := GetOrCreateTeamsClient(ctx)
	if err != nil {
		log.Error("Failed to create Teams client", "error", err)
		return err
	}

	log.Info("Retrieving chats")
	chats, err := teamsClient.Chats.List(ctx)
	if err != nil {
		log.Error("Failed to retrieve chats", "error", err)
		return err
	}

	log.Info("Retrieved chats", "count", len(chats))

	if len(chats) == 0 {
		fmt.Println("No chats found")
		return nil
	}

	fmt.Printf("Found %d chats:\n\n", len(chats))
	for i, chat := range chats {
		fmt.Printf("Chat %d:\n", i+1)
		fmt.Printf("  ID: %s\n", chat.ID)
		if chat.Topic != nil {
			fmt.Printf("  Topic: %s\n", *chat.Topic)
		}
		fmt.Printf("  Type: %s\n", chat.Type)
	}

	return nil
}
