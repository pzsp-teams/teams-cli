package templates

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"text/template"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
)

var htmlTagRegex = regexp.MustCompile(`</?[ibp]>|<br>|<a\s+href="[^"]*">|</a>`)

// TemplateParser handles parsing different messages from supplied template and data
type TemplateParser struct {
	template   *template.Template
	recipients map[string]file_readers.TemplateData
}

// NewMessageParser returns a MessageParser with given config.
// It parses the template and data immediately, storing the parsed objects.
func NewMessageParser(templateReader, dataReader io.Reader, dataParser file_readers.DecodeFunc) (*TemplateParser, error) {
	tmpl, err := readTemplate(templateReader)
	if err != nil {
		// readTemplate already logs and wraps the error
		return nil, err
	}

	recipients := make(map[string]file_readers.TemplateData)
	err = dataParser(dataReader, &recipients)
	if err != nil {
		initializers.Logger.Error(errDataParseFailed.Error(), "error", err)
		return nil, fmt.Errorf("%w: %w", errDataParseFailed, err)
	}
	initializers.Logger.Info("Message data parsed", "recipient_count", len(recipients))

	return &TemplateParser{
		template:   tmpl,
		recipients: recipients,
	}, nil
}

// Parse renders the template for each recipient and returns a map of rendered messages.
// The map keys are recipient names, and values are the fully rendered messages.
func (mp *TemplateParser) Parse() (map[string]string, error) {
	messages := make(map[string]string, len(mp.recipients))
	for recipientName, data := range mp.recipients {
		var buf bytes.Buffer
		if err := mp.template.Execute(&buf, data); err != nil {
			initializers.Logger.Error(errTemplateRenderFailed.Error(), "recipient", recipientName, "error", err)
			return nil, fmt.Errorf("%w for recipient %q: %w", errTemplateRenderFailed, recipientName, err)
		}
		messages[recipientName] = processContent(buf.Bytes())
	}

	initializers.Logger.Info("Successfully rendered messages", "total_messages", len(messages))
	return messages, nil
}

func processContent(data []byte) string {
	if htmlTagRegex.Match(data) {
		return string(data)
	}

	// \r can be ignored, won't be rendered by html on teams anyway
	replaced := bytes.ReplaceAll(data, []byte("\n"), []byte("<br>"))
	return string(replaced)
}
