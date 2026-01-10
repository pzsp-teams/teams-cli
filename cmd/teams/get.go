package teams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
	"github.com/pzsp-teams/lib/models"
)

type getFlags struct {
	team string
}

// newGetCommand creates the teams get command
func newGetCommand() *cobra.Command {
	flags := &getFlags{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a specific team",
		Long:  `Retrieve and display details of a specific Microsoft Teams team by its ID or display name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "ID or display name of the team to get details for")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runGet(cmd *cobra.Command, flags *getFlags) error {
	c, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	t, err := c.Teams.Get(cmd.Context(), flags.team)
	if err != nil {
		return err
	}

	printTeamDetails(t)
	return nil
}

func printTeamDetails(t *models.Team) {
	fmt.Printf("Team Details:\n")
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Display Name: %s\n", t.DisplayName)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Is Archived: %t\n", t.IsArchived)
	if t.Visibility != nil {
		fmt.Printf("Visibility: %s\n", *t.Visibility)
	}
}
