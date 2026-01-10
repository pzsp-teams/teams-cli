package messages

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the channels messages command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Manage channel messages",
		Long:  `Commands for retrieving and managing messages in channels`,
	}

	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newReplyCommand())

	return cmd
}
