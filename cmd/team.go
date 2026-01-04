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
	tRef string
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
}

func runTeamList(cmd *cobra.Command, args []string) error {
	c, err := GetOrCreateTeamsClient(cmd.Context())
	if err != nil {
		return err
	}
	ts, err := c.Client.Teams.ListMyJoined(cmd.Context())
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
	t, err := c.Client.Teams.Get(cmd.Context(), tRef)
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
	_, err = c.Client.Teams.CreateFromTemplate(cmd.Context(), newTeamDisplayName, newTeamDescription, nil)
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
