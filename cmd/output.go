package cmd

import (
	"fmt"

	"github.com/pzsp-teams/lib/models"
	chatsretriever "github.com/pzsp-teams/cli/internal/chats/retriever"
)

// printMessages prints retrieved messages in a readable format
func printMessages(messages []*models.Message, sourceType string) {
	if len(messages) == 0 {
		fmt.Printf("No messages found in the specified time range\n")
		return
	}

	fmt.Printf("Retrieved %d messages from %ss:\n\n", len(messages), sourceType)

	for i, msg := range messages {
		if msg.Content == "" {
			continue
		}

// printChatMessages prints retrieved chat messages in a readable format
func printChatMessages(messages []*chatsretriever.ChatMessageWithContext) {
	if len(messages) == 0 {
		fmt.Printf("No messages found in the specified time range\n")
		return
	}

	fmt.Printf("Retrieved %d messages from chats:\n\n", len(messages))

	for i, msgCtx := range messages {
		msg := msgCtx.Message
		if msg.Content == "" {
			continue
		}

		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  Chat: %s\n", msgCtx.ChatName)
		if msgCtx.ChatType != "" {
			fmt.Printf("  Type: %s\n", msgCtx.ChatType)
		}
		fmt.Printf("  ID: %s\n", msg.ID)
		fmt.Printf("  From: %s\n", msg.From.DisplayName)
		if msg.From.UserID != "" {
			fmt.Printf("  User ID: %s\n", msg.From.UserID)
		}
		fmt.Printf("  Created: %s\n", msg.CreatedDateTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Content:\n%s\n", msg.Content)
		fmt.Printf("\n")
	}
}
