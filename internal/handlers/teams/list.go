package teams

import (
	"context"
	"fmt"
	"io"

	"github.com/pzsp-teams/teams-cli/internal/client"
)

// ListTeams handles the listing of teams.
// It is purely business logic, decoupled from Cobra/TUI.
func ListTeams(ctx context.Context, w io.Writer, flags map[string]any) (any, error) {
	// Initialize client (assuming global init or passed in context,
	// but here we use the internal client helper for now)
	c, err := client.GetInstance()
	if err != nil {
		return nil, err
	}

	ts, err := c.Teams.ListMyJoined(ctx)
	if err != nil {
		return nil, err
	}

	if len(ts) == 0 {
		_, _ = fmt.Fprintln(w, "No teams found.")
		return nil, nil
	}

	// For now, write to stdout directly.
	// Ideally, this should take an io.Writer for better testability.
	_, _ = fmt.Fprintln(w, "Teams:")
	for _, t := range ts {
		state := ""
		if t.IsArchived {
			state = " (Archived)"
		}
		_, _ = fmt.Fprintf(w, "- %s%s (ID: %s)\n", t.DisplayName, state, t.ID)
	}

	return ts, nil
}
