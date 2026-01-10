package teams

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the teams parent command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Simple team operations",
		Long:  `Commands for interacting with Microsoft Teams teams`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCreateSingleCommand())
	cmd.AddCommand(newArchiveCommand())
	cmd.AddCommand(newUnarchiveCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}
