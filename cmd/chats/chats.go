package chats

import (
	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/chats/messages"
)

// NewCommand creates the chats parent command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "Manage chat operations",
		Long:  `Commands for interacting with Microsoft Teams chats`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(messages.NewCommand())

	return cmd
}
