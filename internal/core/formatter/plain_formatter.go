package formatter

import (
	"html"
	"strings"
)

type plainTextFormatter struct{}

// NewPlainTextFormatter creates a formatter that converts HTML to plain text
func NewPlainTextFormatter() Formatter {
	return &plainTextFormatter{}
}

// Format converts HTML content to plain text using regex substitutions
func (f *plainTextFormatter) Format(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	content := htmlContent

	content = html.UnescapeString(content)

	// Convert non-breaking spaces to regular spaces
	content = strings.ReplaceAll(content, "\u00a0", " ")

	content = mentionPattern.ReplaceAllString(content, "@$1")

	content = linkPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := linkPattern.FindStringSubmatch(match)
		if len(submatches) == 3 {
			url := submatches[1]
			text := submatches[2]
			if url == text {
				return text
			}
			return text + " (" + url + ")"
		}
		return match
	})

	content = brPattern.ReplaceAllString(content, "\n")

	content = reduceConsecutiveNewlines(content, 1)

	content = boldPattern.ReplaceAllString(content, "$1")
	content = italicPattern.ReplaceAllString(content, "$1")
	content = underlinePattern.ReplaceAllString(content, "$1")
	content = strikePattern.ReplaceAllString(content, "$1")

	content = paragraphPattern.ReplaceAllString(content, "\n\n$1\n\n")

	content = htmlTagPattern.ReplaceAllString(content, "")

	content = cleanupWhitespace(content, 2)

	return content
}
