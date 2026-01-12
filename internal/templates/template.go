package templates

import (
	"fmt"
	"io"
	"text/template"

	"github.com/pzsp-teams/teams-cli/internal/initializers"
)

// readTemplate reads template content from r and returns a parsed text/template.
// The template uses Go's text/template syntax with {{.placeholder}} format.
// Returns an error if reading fails or if the template syntax is invalid.
// Templates are configured to return an error if any placeholder is missing from the data.
func readTemplate(r io.Reader) (*template.Template, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		initializers.Logger.Error(errTemplateReadFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errTemplateReadFailed, err)
	}

	tmpl, err := template.New("message").Option("missingkey=error").Parse(string(content))
	if err != nil {
		initializers.Logger.Error(errTemplateParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errTemplateParseFailed, err)
	}

	return tmpl, nil
}
