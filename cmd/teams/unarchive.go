package teams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
)

type unarchiveFlags struct {
	team string
}

// newUnarchiveCommand creates the teams unarchive command
func newUnarchiveCommand() *cobra.Command {
	flags := &unarchiveFlags{}

	cmd := &cobra.Command{
		Use:   "unarchive",
		Short: "Unarchive a team",
		Long:  `Unarchive a Microsoft Teams team by its ID or display name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnarchive(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "ID or display name of the team to unarchive")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runUnarchive(cmd *cobra.Command, flags *unarchiveFlags) error {
	c, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	err = c.Teams.Unarchive(cmd.Context(), flags.team)
	if err != nil {
		return err
	}

	fmt.Println("Team unarchive initiated. The team will be unarchived shortly.")
	return nil
}
