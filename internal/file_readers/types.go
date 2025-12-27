package file_readers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
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
)

// DecodeFunc defines a function type for decoding data from an io.Reader into a provided structure
type DecodeFunc func(r io.Reader, v any) error

// DecodeJSON decodes JSON data from the reader into the provided structure
func DecodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, JSON, err)
	}
	return nil
}

// DecodeYAML decodes YAML data from the reader into the provided structure
func DecodeYAML(r io.Reader, v any) error {
	if err := yaml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, YAML, err)
	}
	return nil
}

// DecodeTOML decodes TOML data from the reader into the provided structure
func DecodeTOML(r io.Reader, v any) error {
	if _, err := toml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, TOML, err)
	}
	return nil
}
