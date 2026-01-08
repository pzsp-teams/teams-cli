package formatters

import (
	"io"
	"time"
)

// Format represents different output formats for message content
type Format string

const (
	// FormatPlainText converts HTML to plain text
	FormatPlainText Format = "plain"
	// FormatMarkdown converts HTML to Markdown (future)
	FormatMarkdown Format = "markdown"
)

// ContextItem represents a key-value pair for message context (e.g., Team, Channel)
type ContextItem struct {
	Label string
	Value string
}

// MessageView represents the generic data needed to display a message
type MessageView struct {
	ID        string
	Author    string
	Timestamp time.Time
	Content   string
	Context   []ContextItem
}

// Formatter provides content formatting and writing capabilities
type Formatter interface {
	// Format converts raw (HTML) content into the target format string
	Format(htmlContent string) string
	// WriteMessages formats and writes a collection of messages to the writer
	WriteMessages(w io.Writer, messages []*MessageView) error
}
