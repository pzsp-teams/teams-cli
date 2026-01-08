package formatter

// Format represents different output formats for message content
type Format string

const (
	// FormatPlainText converts HTML to plain text
	FormatPlainText Format = "plain"
	// FormatMarkdown converts HTML to Markdown (future)
	FormatMarkdown Format = "markdown"
)

// Formatter provides content formatting capabilities
type Formatter interface {
	Format(htmlContent string) string
}
