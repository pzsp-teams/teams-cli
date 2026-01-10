package teams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
)

type deleteFlags struct {
	team string
}

// newDeleteCommand creates the teams delete command
func newDeleteCommand() *cobra.Command {
	flags := &deleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a team",
		Long:  `Delete a Microsoft Teams team by its ID or display name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "ID or display name of the team to delete")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runDelete(cmd *cobra.Command, flags *deleteFlags) error {
	c, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	err = c.Teams.Delete(cmd.Context(), flags.team)
	if err != nil {
		return err
	}

	fmt.Println("Team deletion initiated. The team will be deleted shortly.")
	return nil
}
