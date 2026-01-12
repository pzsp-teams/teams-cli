package tui

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pzsp-teams/teams-cli/app"
)

// CommandExecutor executes CLI commands from TUI
type CommandExecutor struct {
	ctx context.Context
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(ctx context.Context) *CommandExecutor {
	return &CommandExecutor{ctx: ctx}
}

// ExecuteCommand executes an app handler
func (e *CommandExecutor) ExecuteCommand(def *app.CommandDef, flags map[string]any) (output string, data any, err error) {
	if def.Handler == nil {
		return "", nil, fmt.Errorf("no handler defined for command: %s", def.Use)
	}

	if err := app.ValidateFlags(flags, def.Flags); err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	data, err = def.Handler(e.ctx, &buf, flags)
	output = buf.String()
	return output, data, err
}
