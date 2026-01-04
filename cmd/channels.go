package cmd

import (
	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage channels operations",
	Long:  `Commands for interacting with Microsoft Teams channels`,
}

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(channelsSendCmd)
	channelsCmd.AddCommand(createChannelsCmd)
}
