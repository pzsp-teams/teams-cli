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

// NopWriteCloser returns an io.WriteCloser that wraps the given io.Writer
// and has a no-op Close method.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// GetDest returns an io.WriteCloser for the given filename
// If filename is empty, returns stdout wrapped in a no-op closer
// Otherwise opens/creates the file for appending
func GetDest(filename string) (io.WriteCloser, error) {
	if filename == "" {
		return &nopWriteCloser{os.Stdout}, nil
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", filename, err)
	}
	return file, nil
}

// GetFormatter returns a formatter based on the plain and markdown flags
// Returns plain text formatter by default
func GetFormatter(format string) formatters.Formatter {
	switch format {
	case "markdown":
		return formatters.NewMarkdownFormatter()
	default:
		return formatters.NewPlainTextFormatter()
	}
}
