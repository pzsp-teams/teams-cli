package tui

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/cli/app"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run(ctx context.Context, registry []app.CommandDef, startPath string) error {
	p := tea.NewProgram(initialModel(ctx, registry, startPath), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}