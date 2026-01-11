package teams

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
)

// DeleteTeam handles deleting a team.
func DeleteTeam(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = c.Teams.Delete(ctx, team)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(w, "Team deletion initiated. The team will be deleted shortly.")
	return nil, nil
}
