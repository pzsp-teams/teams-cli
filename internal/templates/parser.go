package templates

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/cli/internal/initializers"
	"github.com/pzsp-teams/lib/models"
)

var (
	htmlTagRegex = regexp.MustCompile(`</?[ibp]>|<br>|<a\s+href="[^"]*">|</a>`)
	mentionRegex = regexp.MustCompile(`@@[\s]*@[\s]*([^\s]+)[\s]*@@`)
)

// TemplateParser handles parsing different messages from supplied template and data
type TemplateParser struct {
	template   *template.Template
	recipients map[string]TemplateData
}

// NewMessageParser returns a MessageParser with given config.
// It parses the template and data immediately, storing the parsed objects.
func NewMessageParser(templateReader, dataReader io.Reader, dataParser file_readers.DecodeFunc) (*TemplateParser, error) {
	tmpl, err := readTemplate(templateReader)
	if err != nil {
		// readTemplate already logs and wraps the error
		return nil, err
	}

	recipients := make(map[string]TemplateData)
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

// ExtractMentions extracts user mention patterns from content and returns raw mention strings
func ExtractMentions(content string) []string {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	rawMentions := []string{}

	for _, match := range matches {
		if len(match) > 1 {
			raw := strings.TrimSpace(match[1])
			if raw != "" && !seen[raw] {
				rawMentions = append(rawMentions, raw)
				seen[raw] = true
			}
		}
	}

	return rawMentions
}

// ReplaceMentions replaces user mention patterns with <at> tags using the provided mentions in order
// Each occurrence gets a unique AtID, mentions list will contain duplicates for repeated mentions
func ReplaceMentions(content string, mentions []models.Mention) string {
	if len(mentions) == 0 {
		return content
	}

	processed := content

	for _, mention := range mentions {
		match := mentionRegex.FindStringIndex(processed)
		if match == nil {
			break
		}

		atTag := fmt.Sprintf(`<at id="%d">%s</at>`, mention.AtID, mention.Text)
		processed = processed[:match[0]] + atTag + processed[match[1]:]
	}

	return processed
}
