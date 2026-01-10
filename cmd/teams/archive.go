package teams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pzsp-teams/cli/cmd/common"
)

type archiveFlags struct {
	team        string
	spoReadOnly bool
}

// newArchiveCommand creates the teams archive command
func newArchiveCommand() *cobra.Command {
	flags := &archiveFlags{}

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive a team",
		Long:  `Archive a Microsoft Teams team by its ID or display name.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive(cmd, flags)
		},
	}

	cmd.Flags().StringVar(&flags.team, "team", "", "ID or display name of the team to archive")
	cmd.Flags().BoolVar(&flags.spoReadOnly, "spo-read-only", false, "Set SharePoint Online site to read-only mode when archiving the team")
	if err := cmd.MarkFlagRequired("team"); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to mark team flag as required: %v\n", err)
	}

	return cmd
}

func runArchive(cmd *cobra.Command, flags *archiveFlags) error {
	c, err := common.GetTeamsClient(cmd)
	if err != nil {
		return err
	}

	err = c.Teams.Archive(cmd.Context(), flags.team, flags.spoReadOnly)
	if err != nil {
		return err
	}

	fmt.Println("Team archive initiated. The team will be archived shortly.")
	return nil
}
