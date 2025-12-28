package file_readers

import (
	"io"
)

// TemplateData represents placeholder values for a single message recipient
type TemplateData map[string]string

// Formats represents supported file formats for decoding
type Formats string

// Supported file formats
const (
	JSON Formats = "json"
	YAML Formats = "yaml"
	TOML Formats = "toml"
	CSV  Formats = "csv"
)

// DecodeFunc defines a function type for decoding data from an io.Reader into a provided structure
type DecodeFunc func(r io.Reader, v any) error
