package cmd

import (
	"github.com/spf13/cobra"
)

var chatsCmd = &cobra.Command{
	Use:   "chats",
	Short: "Manage chat operations",
	Long:  `Commands for interacting with Microsoft Teams chats`,
}

func init() {
	rootCmd.AddCommand(chatsCmd)
	chatsCmd.AddCommand(chatsSendCmd)
}
