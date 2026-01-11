package teams

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
	"github.com/pzsp-teams/lib/models"
)

// GetTeam handles retrieving team details.
func GetTeam(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	t, err := c.Teams.Get(ctx, team)
	if err != nil {
		return nil, err
	}

	printTeamDetails(w, t)
	return t, nil
}

func printTeamDetails(w io.Writer, t *models.Team) {
	fmt.Fprintln(w, "Team Details:")
	fmt.Fprintf(w, "ID: %s\n", t.ID)
	fmt.Fprintf(w, "Display Name: %s\n", t.DisplayName)
	fmt.Fprintf(w, "Description: %s\n", t.Description)
	fmt.Fprintf(w, "Is Archived: %t\n", t.IsArchived)
	if t.Visibility != nil {
		fmt.Fprintf(w, "Visibility: %s\n", *t.Visibility)
	}
}
