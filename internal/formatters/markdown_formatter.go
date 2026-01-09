package formatters

import (
	"fmt"
	"html"
	"io"
	"strings"
	"time"
)

type markdownFormatter struct{}

// NewMarkdownFormatter creates a formatter that converts HTML to plain text
func NewMarkdownFormatter() Formatter {
	return &markdownFormatter{}
}

// Format converts HTML content to plain text using regex substitutions
func (f *markdownFormatter) Format(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	content := htmlContent
	content = html.UnescapeString(content)
	content = strings.ReplaceAll(content, "\u00a0", " ")

	content = paragraphPattern.ReplaceAllString(content, "\n\n$1\n\n")

	content = multipleBrPattern.ReplaceAllString(content, "<br>")

	content = brPattern.ReplaceAllString(content, " \\\n")

	content = reduceConsecutiveNewlines(content, 2)

	content = strings.ReplaceAll(content, " \\\n\n", "\n\n")

	content = boldPattern.ReplaceAllString(content, "**$1**")
	content = italicPattern.ReplaceAllString(content, "*$1*")
	content = strikePattern.ReplaceAllString(content, "--$1--")
	content = italicEmptyPattern.ReplaceAllString(content, "")
	content = boldEmptyPattern.ReplaceAllString(content, "")

	content = mentionPattern.ReplaceAllString(content, "@$1")

	content = unwantedTagPattern.ReplaceAllString(content, "")

	content = linkPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := linkPattern.FindStringSubmatch(match)
		if len(submatches) == 3 {
			url := submatches[1]
			text := submatches[2]
			if url == text {
				return "<" + text + "/>"
			}
			return "[" + text + "](" + url + ")"
		}
		return match
	})

	content = cleanupWhitespace(content, 2)

	return content
}

// WriteMessages formats and writes a collection of messages to the writer
func (f *markdownFormatter) WriteMessages(w io.Writer, messages []*MessageView) error {
	for _, msg := range messages {
		if _, err := fmt.Fprintf(w, "### From %s\n\n", msg.Author); err != nil {
			return err
		}

		for _, ctx := range msg.Context {
			if _, err := fmt.Fprintf(w, "**%s:** %s \\\n", ctx.Label, ctx.Value); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "**Date:** %s \\\n", msg.Timestamp.Format(time.RFC822)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "**ID:** %s", msg.ID); err != nil {
			return err
		}

		if _, err := fmt.Fprint(w, "\n\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, f.Format(msg.Content)); err != nil {
			return err
		}

		if _, err := fmt.Fprint(w, "\n---\n\n"); err != nil {
			return err
		}
	}
	return nil
}
