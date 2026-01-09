package formatters

import (
	"regexp"
	"strings"
)

// Patterns for supported HTML tags
var (
	// Mention: <at id="X">Name</at>
	mentionPattern = regexp.MustCompile(`<at\s+id="[^"]*">([^<]+)</at>`)

	// Links: <a href="url">text</a> - (?s) enables DOTALL mode so . matches newlines
	linkPattern = regexp.MustCompile(`(?s)<a\s+href="([^"]+)">(.+?)</a>`)

	// Bold: <b>text</b> or <strong>text</strong>
	boldPattern      = regexp.MustCompile(`<(?:b|strong)>(.+?)</(?:b|strong)>`)
	boldEmptyPattern = regexp.MustCompile(`<(?:b|strong)></(?:b|strong)>`)

	// Italic: <i>text</i> or <em>text</em>
	italicPattern      = regexp.MustCompile(`<(?:i|em)>(.+?)</(?:i|em)>`)
	italicEmptyPattern = regexp.MustCompile(`<(?:i|em)></(?:i|em)>`)

	// Underline: <u>text</u>
	underlinePattern = regexp.MustCompile(`<u>(.+?)</u>`)

	// Strikethrough: <s>text</s> or <strike>text</strike>
	strikePattern = regexp.MustCompile(`<(?:s|strike)>(.+?)</(?:s|strike)>`)

	// Paragraphs: <p>text</p> - (?s) enables DOTALL mode so . matches newlines
	paragraphPattern = regexp.MustCompile(`(?s)<p>(.*?)</p>`)

	// Line breaks: <br> or <br/>
	brPattern = regexp.MustCompile(`<br\s*/?>`)

	// Multiple consecutive line breaks (for collapsing)
	multipleBrPattern = regexp.MustCompile(`(?:<br\s*/?>(?:\s*<br\s*/?>)+)`)

	// Any remaining HTML tags (for cleanup) - must start with a letter
	htmlTagPattern = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)

	// Unwanted HTML tags to remove (unclosed paragraphs, divs, spans, etc.)
	unwantedTagPattern = regexp.MustCompile(`</?(?:p|div|span)[^>]*>`)
)

// reduceConsecutiveNewlines reduces consecutive newlines to a maximum count
func reduceConsecutiveNewlines(s string, maxNewlines int) string {
	if maxNewlines < 1 {
		maxNewlines = 1
	}

	targetPattern := strings.Repeat("\n", maxNewlines+1)
	replacement := strings.Repeat("\n", maxNewlines)

	for strings.Contains(s, targetPattern) {
		s = strings.ReplaceAll(s, targetPattern, replacement)
	}

	return s
}

// cleanupWhitespace removes excessive whitespace while preserving intentional formatting
func cleanupWhitespace(s string, maxNewlines int) string {
	for strings.Contains(s, " \n") || strings.Contains(s, "\t\n") {
		s = strings.ReplaceAll(s, " \n", "\n")
		s = strings.ReplaceAll(s, "\t\n", "\n")
	}

	s = strings.TrimLeft(s, " \t\n\r")

	s = strings.TrimRight(s, " \t")

	s = reduceConsecutiveNewlines(s, maxNewlines)

	s = strings.TrimRight(s, "\n")

	s = strings.TrimRight(s, "\\")

	s = strings.TrimRight(s, " \t")

	if strings.TrimSpace(s) == "" {
		return ""
	}

	return s
}
