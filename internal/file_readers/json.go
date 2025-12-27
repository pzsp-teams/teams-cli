package file_readers

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pzsp-teams/cli/internal/initializers"
)

// JSONParser implements Parser for JSON format
type JSONParser struct{}

// Parse reads JSON-formatted message data
func (p *JSONParser) Parse(r io.Reader) (map[string]TemplateData, error) {
	var messages map[string]TemplateData
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&messages); err != nil {
		initializers.Logger.Error(errJSONDecodeFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errJSONDecodeFailed, err)
	}
	return messages, nil
}
