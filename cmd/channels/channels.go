package channels

import (
	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/channels/messages"
)

// NewCommand creates the channels parent command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Manage channels operations",
		Long:  `Commands for interacting with Microsoft Teams channels`,
	}

	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(messages.NewCommand())

	return cmd
}
