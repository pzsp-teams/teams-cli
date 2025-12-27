package file_readers

import "io"

// TemplateData represents placeholder values for a single message recipient
type TemplateData map[string]string

// Parser defines the interface for parsing message data from different formats
type Parser interface {
	// Parse reads and parses message data from r
	Parse(r io.Reader) (map[string]TemplateData, error)
}
