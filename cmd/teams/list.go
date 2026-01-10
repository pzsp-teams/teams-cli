package teams

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
)

// newListCommand creates the teams list command
func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all teams the user is a member of",
		Long:  `Retrieve and display a list of all Microsoft Teams that the authenticated user is a member of.`,
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	c, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	ts, err := c.Teams.ListMyJoined(cmd.Context())
	if err != nil {
		return err
	}

	if len(ts) == 0 {
		cmd.Println("No teams found.")
		return nil
	}

	fmt.Println("Teams:")
	for _, t := range ts {
		state := ""
		if t.IsArchived {
			state = " (Archived)"
		}
		fmt.Printf("- %s%s (ID: %s)\n", t.DisplayName, state, t.ID)
	}

	return nil
}
