package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}
