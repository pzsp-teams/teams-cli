package tui

import (
	"github.com/pzsp-teams/teams-cli/app"
)

// navigationState tracks current command and navigation history
type navigationState struct {
	current *app.CommandDef
	history []*app.CommandDef
}

func newNavigationState() *navigationState {
	return &navigationState{
		current: nil, // Root
		history: []*app.CommandDef{},
	}
}

func (n *navigationState) navigateTo(def *app.CommandDef) {
	n.history = append(n.history, n.current)
	n.current = def
}

func (n *navigationState) goBack() bool {
	if len(n.history) == 0 {
		return false
	}
	n.current = n.history[len(n.history)-1]
	n.history = n.history[:len(n.history)-1]
	return true
}

func (n *navigationState) getCurrentCommand() *app.CommandDef {
	return n.current
}
