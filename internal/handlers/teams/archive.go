package teams

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/client"
)

// ArchiveTeam handles archiving a team.
func ArchiveTeam(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	team, _ := flags["team"].(string)
	spoReadOnly, _ := flags["spo-read-only"].(bool)

	c, err := client.GetOrCreateInstance(ctx)
	if err != nil {
		return nil, err
	}

	err = c.Teams.Archive(ctx, team, spoReadOnly)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintln(w, "Team archive initiated. The team will be archived shortly.")
	return nil, nil
}
