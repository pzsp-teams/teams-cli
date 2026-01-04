package formatter

import (
	"regexp"
	"strings"
)

// Patterns for supported HTML tags
var (
	// Mention: <at id="X">Name</at>
	mentionPattern = regexp.MustCompile(`<at\s+id="[^"]*">([^<]+)</at>`)

	// Links: <a href="url">text</a>
	linkPattern = regexp.MustCompile(`<a\s+href="([^"]+)">(.+?)</a>`)

	// Bold: <b>text</b> or <strong>text</strong>
	boldPattern = regexp.MustCompile(`<(?:b|strong)>(.+?)</(?:b|strong)>`)

	// Italic: <i>text</i> or <em>text</em>
	italicPattern = regexp.MustCompile(`<(?:i|em)>(.+?)</(?:i|em)>`)

	// Underline: <u>text</u>
	underlinePattern = regexp.MustCompile(`<u>(.+?)</u>`)

	// Strikethrough: <s>text</s> or <strike>text</strike>
	strikePattern = regexp.MustCompile(`<(?:s|strike)>(.+?)</(?:s|strike)>`)

	// Paragraphs: <p>text</p>
	paragraphPattern = regexp.MustCompile(`<p>(.+?)</p>`)

	// Line breaks: <br> or <br/>
	brPattern = regexp.MustCompile(`<br\s*/?>`)

	// Any remaining HTML tags (for cleanup) - must start with a letter
	htmlTagPattern = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
)

// cleanupWhitespace removes excessive whitespace while preserving intentional formatting
func cleanupWhitespace(s string) string {
	// Remove spaces/tabs before newlines
	for strings.Contains(s, " \n") || strings.Contains(s, "\t\n") {
		s = strings.ReplaceAll(s, " \n", "\n")
		s = strings.ReplaceAll(s, "\t\n", "\n")
	}

	// Trim leading whitespace
	s = strings.TrimLeft(s, " \t\n\r")

	// Trim trailing spaces/tabs only
	s = strings.TrimRight(s, " \t")

	// Limit consecutive newlines to maximum of 2
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}

	if strings.TrimSpace(s) == "" {
		return ""
	}

	return s
}
