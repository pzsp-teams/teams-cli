package cmd

import (
	"github.com/spf13/cobra"
)

var channelsMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Manage channel messages",
	Long:  `Commands for retrieving and managing messages in channels`,
}

func init() {
	channelsMessagesCmd.AddCommand(channelsMessagesGetCmd)
}
