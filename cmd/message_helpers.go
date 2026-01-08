package cmd

import (
	"fmt"
	"github.com/pzsp-teams/cli/internal/formatters"
)

type formatFlags struct {
	Plain    bool
	MarkDown bool
}

func (f *formatFlags) getFormatter() (formatters.Formatter, error) {
	count := 0
	var format formatters.Formatter

	if f.MarkDown {
		count++
		format = formatters.NewMarkdownFormatter()
	}
	if f.Plain {
		count++
		format = formatters.NewPlainTextFormatter()
	}
	if count == 0 {
		format = formatters.NewPlainTextFormatter()
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple formats specified, please choose only one")
	}

	return format, nil
}
