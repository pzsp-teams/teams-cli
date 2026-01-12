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

	c, err := client.GetInstance()
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
	_, _ = fmt.Fprintln(w, "Team Details:")
	_, _ = fmt.Fprintf(w, "ID: %s\n", t.ID)
	_, _ = fmt.Fprintf(w, "Display Name: %s\n", t.DisplayName)
	_, _ = fmt.Fprintf(w, "Description: %s\n", t.Description)
	_, _ = fmt.Fprintf(w, "Is Archived: %t\n", t.IsArchived)
	if t.Visibility != nil {
		_, _ = fmt.Fprintf(w, "Visibility: %s\n", *t.Visibility)
	}
}
