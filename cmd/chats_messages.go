package cmd

import (
	"github.com/spf13/cobra"
)

var chatsMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Manage chat messages",
	Long:  `Commands for retrieving and managing messages in chats`,
}

func init() {
	chatsMessagesCmd.AddCommand(chatsMessagesGetCmd)
}
