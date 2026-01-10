package common

import (
	"fmt"
	"io"
	"os"

	"github.com/pzsp-teams/cli/internal/formatters"
)

type nopWriteCloser struct {
	io.Writer
}

// Close implements io.Closer for nopWriteCloser (no-op implementation)
func (nc *nopWriteCloser) Close() error {
	return nil
}

// GetDest returns an io.WriteCloser for the given filename
// If filename is empty, returns stdout wrapped in a no-op closer
// Otherwise opens/creates the file for appending
func GetDest(filename string) io.WriteCloser {
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

// GetFormatter returns a formatter based on the plain and markdown flags
// Returns plain text formatter by default
func GetFormatter(plain, markdown bool) (formatters.Formatter, error) {
	count := 0
	var format formatters.Formatter

	if markdown {
		count++
		format = formatters.NewMarkdownFormatter()
	}
	if plain {
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
