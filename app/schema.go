package app

import (
	"context"
	"fmt"
	"io"
	"slices"
)

// InputType represents the type of UI input required
type InputType int

const (
	InputNone InputType = iota
	InputString
	InputLongString
	InputInt
	InputBool
	InputPassword
	InputFile
	InputDate
	InputChoice
)

// OutputType represents the type of data returned by the handler
type OutputType int

const (
	OutputText        OutputType = iota
	OutputList                   // Generic list
	OutputMessageList            // List of messages (interactive)
)

// FlagDef defines a command argument/flag
type FlagDef struct {
	Name          string
	Shorthand     string
	Usage         string
	Type          InputType
	DefaultVal    any
	Required      bool
	Options       []string // For InputChoice - valid choices
	RequiresFlags []string // Flags that must be set with this one
	ConflictsWith []string // Flags that cannot be used with this one
}

// HandlerFunc is the logic that runs for a command.
// It receives context, a writer for text output, and flag values.
// It returns a structured result and an error.
type HandlerFunc func(ctx context.Context, w io.Writer, flags map[string]any) (any, error)

// CommandDef defines a single command in the application
type CommandDef struct {
	Use         string
	Short       string
	Long        string
	Flags       []FlagDef
	Handler     HandlerFunc
	OutputType  OutputType
	SubCommands []CommandDef
}

// Registry holds the root commands of the application
var Registry = []CommandDef{}

// ValidateFlags checks flag dependencies and conflicts
func ValidateFlags(flags map[string]any, defs []FlagDef) error {
	for i := range defs {
		def := &defs[i]
		val := flags[def.Name]

		if !isSet(flags, def.Name) {
			continue
		}

		if def.Type == InputChoice {
			strVal, ok := val.(string)
			if !ok {
				continue
			}
			if !contains(def.Options, strVal) {
				return fmt.Errorf("invalid value for --%s: must be one of %v", def.Name, def.Options)
			}
		}

		for _, req := range def.RequiresFlags {
			if !isSet(flags, req) {
				return fmt.Errorf("--%s requires --%s", def.Name, req)
			}
		}

		for _, conflict := range def.ConflictsWith {
			if isSet(flags, conflict) {
				return fmt.Errorf("--%s cannot be used with --%s", def.Name, conflict)
			}
		}
	}
	return nil
}

func isSet(flags map[string]any, name string) bool {
	val, exists := flags[name]
	if !exists {
		return false
	}

	switch v := val.(type) {
	case string:
		return v != ""
	case bool:
		return v
	case int:
		return true
	default:
		return val != nil
	}
}

func contains(slice []string, val string) bool {
	return slices.Contains(slice, val)
}
