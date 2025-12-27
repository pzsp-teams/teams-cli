package templates

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
)

// Registry manages available parsers for different file formats
type Registry struct {
	parsers map[string]file_readers.DecodeFunc
}

// NewParserRegistry creates a new registry with default parsers
func NewParserRegistry() *Registry {
	registry := &Registry{
		parsers: make(map[string]file_readers.DecodeFunc),
	}

	registry.Register("json", file_readers.DecodeJSON)
	registry.Register("yaml", file_readers.DecodeYAML)
	registry.Register("yml", file_readers.DecodeYAML)
	registry.Register("toml", file_readers.DecodeTOML)

	initializers.Logger.Info("Parser registry initialized", "supported_formats", registry.SupportedFormats())
	return registry
}

// Register adds a parser for a specific file extension
func (r *Registry) Register(ext string, parser file_readers.DecodeFunc) {
	normalizedExt := strings.ToLower(ext)
	r.parsers[normalizedExt] = parser
}

// GetParser returns a parser for the given file extension
func (r *Registry) GetParser(filename string) (file_readers.DecodeFunc, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")

	parser, ok := r.parsers[ext]
	if !ok {
		initializers.Logger.Warn(errNoParserRegistered.Error(), "extension", ext, "supported_formats", r.SupportedFormats())
		return nil, fmt.Errorf("%w: .%s", errNoParserRegistered, ext)
	}

	return parser, nil
}

// SupportedFormats returns a list of supported file extensions
func (r *Registry) SupportedFormats() []string {
	formats := make([]string, 0, len(r.parsers))
	for ext := range r.parsers {
		formats = append(formats, ext)
	}
	return formats
}
