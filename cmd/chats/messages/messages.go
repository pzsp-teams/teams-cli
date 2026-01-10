package messages

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the chats messages NewCommand
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Manage chat messages",
		Long:  `Commands for retrieving and sending messages in chats`,
	}

	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSendCommand())

	return cmd
}
