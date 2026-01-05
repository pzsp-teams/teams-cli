package cmd

import (
	"fmt"

	channelsretriever "github.com/pzsp-teams/cli/internal/channels/retriever"
	chatsretriever "github.com/pzsp-teams/cli/internal/chats/retriever"
)

// printChannelMessages prints retrieved channel messages in a readable format
func printChannelMessages(messages []*channelsretriever.ChannelMessageWithContext) {
	if len(messages) == 0 {
		fmt.Printf("No messages found in the specified time range\n")
		return
	}

	fmt.Printf("Retrieved %d messages from channels:\n\n", len(messages))

	for i, msgCtx := range messages {
		msg := msgCtx.Message
		if msg.Content == "" {
			continue
		}

		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  Team: %s\n", msgCtx.TeamName)
		fmt.Printf("  Channel: %s\n", msgCtx.ChannelName)
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
