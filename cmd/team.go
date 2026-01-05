package cmd

import (
	"fmt"

	"github.com/pzsp-teams/lib/models"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Simple team operations",
	Long:  `Commands for simple team operations on Microsoft Teams`,
}

var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams the user is a member of",
	Long:  `Retrieve and display a list of all Microsoft Teams that the authenticated user is a member of.`,
	RunE:  runTeamList,
}

var (
	tRef        string
	spoReadOnly bool
)

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamGetCmd)
	teamGetCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to get details for")
	if err := teamGetCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamCreateCmd)
	teamCreateCmd.Flags().StringVar(&newTeamDisplayName, "name", "", "Display name of the new team")
	teamCreateCmd.Flags().StringVar(&newTeamDescription, "description", "", "Description of the new team")
	if err := teamCreateCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Sprintf("failed to mark name flag as required: %v", err))
	}

	teamCmd.AddCommand(teamArchiveCmd)
	teamArchiveCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to archive")
	teamArchiveCmd.Flags().BoolVar(&spoReadOnly, "spo-read-only", false, "Set SharePoint Online site to read-only mode when archiving the team")
	if err := teamArchiveCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamUnarchiveCmd)
	teamUnarchiveCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to unarchive")
	if err := teamUnarchiveCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}

	teamCmd.AddCommand(teamDeleteCmd)
	teamDeleteCmd.Flags().StringVar(&tRef, "team", "", "ID or display name of the team to delete")
	if err := teamDeleteCmd.MarkFlagRequired("team"); err != nil {
		panic(fmt.Sprintf("failed to mark team flag as required: %v", err))
	}
}

func runTeamList(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
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

var teamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get details of a specific team",
	Long:  `Retrieve and display details of a specific Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamGet,
}

func runTeamGet(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	t, err := c.Teams.Get(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	printTeamDetails(t)
	return nil
}

var teamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new team",
	Long:  `Create a new Microsoft Teams team with the specified display name and description.`,
	RunE:  runTeamCreate,
}

var (
	newTeamDisplayName string
	newTeamDescription string
)

func runTeamCreate(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	_, err = c.Teams.CreateFromTemplate(cmd.Context(), newTeamDisplayName, newTeamDescription, nil)
	if err != nil {
		return err
	}
	fmt.Println("Team creation initiated. The team will be available once creation is complete.")
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

var teamArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a team",
	Long:  `Archive a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamArchive,
}

func runTeamArchive(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Archive(cmd.Context(), tRef, spoReadOnly)
	if err != nil {
		return err
	}
	fmt.Println("Team archive initiated. The team will be archived shortly.")
	return nil
}

var teamUnarchiveCmd = &cobra.Command{
	Use:   "unarchive",
	Short: "Unarchive a team",
	Long:  `Unarchive a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamUnarchive,
}

func runTeamUnarchive(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Unarchive(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	fmt.Println("Team unarchive initiated. The team will be unarchived shortly.")
	return nil
}

var teamDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a team",
	Long:  `Delete a Microsoft Teams team by its ID or display name.`,
	RunE:  runTeamDelete,
}

func runTeamDelete(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	err = c.Teams.Delete(cmd.Context(), tRef)
	if err != nil {
		return err
	}
	fmt.Println("Team deletion initiated. The team will be deleted shortly.")
	return nil
}
