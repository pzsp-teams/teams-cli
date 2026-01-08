package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/pzsp-teams/cli/internal/formatters"
)

type formatFlags struct {
	Plain    bool
	MarkDown bool
}

type nopWriteCloser struct {
	io.Writer
}

// Close does nothing, implement to wrap stdout
func (nc *nopWriteCloser) Close() error {
	return nil
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

func getDest(filename string) io.WriteCloser {
	if filename == "" {
		return &nopWriteCloser{os.Stdout}
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create %s: %v\n", filename, err)
		os.Exit(1)
	}
	return file
}
