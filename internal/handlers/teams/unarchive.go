package teams

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
)

// UnarchiveTeam handles unarchiving a team.
func UnarchiveTeam(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = c.Teams.Unarchive(ctx, team)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(w, "Team unarchive initiated. The team will be unarchived shortly.")
	return nil, nil
}
