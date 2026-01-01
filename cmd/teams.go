package cmd

import (
	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage Teams operations",
	Long:  `Commands for interacting with Microsoft Teams channels and teams`,
}

func init() {
	rootCmd.AddCommand(teamsCmd)
	teamsCmd.AddCommand(teamsSendCmd)
}
