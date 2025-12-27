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


type Formats string

const (
	JSON Formats = "json"
	YAML Formats = "yaml"
	TOML Formats = "toml"
)

type DecodeFunc func(r io.Reader, v any) error

func DecodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, JSON, err)
	}
	return nil
}

func DecodeYAML(r io.Reader, v any) error {
	if err := yaml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, YAML, err)
	}
	return nil
}

func DecodeTOML(r io.Reader, v any) error {
	if _, err := toml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, TOML, err)
	}
	return nil
}


