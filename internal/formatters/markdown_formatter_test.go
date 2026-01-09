package formatters

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownFormatter_Format(t *testing.T) {
	formatter := NewMarkdownFormatter()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "plain text without HTML",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "text with line break",
			input: "Hello<br>World",
			want:  "Hello\\\nWorld",
		},
		{
			name:  "text with self-closing line break",
			input: "Hello<br/>World",
			want:  "Hello\\\nWorld",
		},
		{
			name:  "text with multiple line breaks",
			input: "Line 1<br>Line 2<br>Line 3",
			want:  "Line 1\\\nLine 2\\\nLine 3",
		},
		{
			name:  "text with bold tags",
			input: "This is <b>bold</b> text",
			want:  "This is **bold** text",
		},
		{
			name:  "text with strong tags",
			input: "This is <strong>important</strong> text",
			want:  "This is **important** text",
		},
		{
			name:  "text with italic tags",
			input: "This is <i>italic</i> text",
			want:  "This is *italic* text",
		},
		{
			name:  "text with em tags",
			input: "This is <em>emphasized</em> text",
			want:  "This is *emphasized* text",
		},
		{
			name:  "text with underline tags",
			input: "This is <u>underlined</u> text",
			want:  "This is <u>underlined</u> text",
		},
		{
			name:  "text with strikethrough tags",
			input: "This is <s>crossed out</s> text",
			want:  "This is --crossed out-- text",
		},
		{
			name:  "simple paragraph",
			input: "<p>This is a paragraph</p>",
			want:  "This is a paragraph",
		},
		{
			name:  "multiple paragraphs",
			input: "<p>First paragraph</p><p>Second paragraph</p>",
			want:  "First paragraph\n\nSecond paragraph",
		},
		{
			name:  "paragraph with formatting",
			input: "<p>This is <b>bold</b> and <i>italic</i></p>",
			want:  "This is **bold** and *italic*",
		},
		{
			name:  "simple mention",
			input: "<at id=\"0\">User Name</at>",
			want:  "@User Name",
		},
		{
			name:  "mention in text",
			input: "Hello <at id=\"1\">John</at>, how are you?",
			want:  "Hello @John, how are you?",
		},
		{
			name:  "multiple mentions",
			input: "Hey <at id=\"0\">Alice</at> and <at id=\"1\">Bob</at>!",
			want:  "Hey @Alice and @Bob!",
		},
		{
			name:  "simple link",
			input: "<a href=\"https://example.com\">Click here</a>",
			want:  "[Click here](https://example.com)",
		},
		{
			name:  "link with same text as URL",
			input: "<a href=\"https://example.com\">https://example.com</a>",
			want:  "<https://example.com/>",
		},
		{
			name:  "link in text",
			input: "Visit <a href=\"https://example.com\">our website</a> for more info",
			want:  "Visit [our website](https://example.com) for more info",
		},
		{
			name:  "nested formatting tags",
			input: "<p><b><i>Bold and italic</i></b></p>",
			want:  "***Bold and italic***",
		},
		{
			name:  "complex nested structure",
			input: "<p>This is <b>bold <i>and italic</i></b> text</p>",
			want:  "This is **bold *and italic*** text",
		},
		{
			name:  "real Teams message example",
			input: "Hello Alice!<br>This is a channel mention <at id=\"0\">General</at><br>This is a personal mention <at id=\"1\">Kamil</at><br>This is a duplicate channel mention <at id=\"2\">General</at><br>This is a team mention <at id=\"3\">pzsp2z1teams</at><br><br>Your order #12345 is ready.<br>Thank you!<br>",
			want:  "Hello Alice!\\\nThis is a channel mention @General\\\nThis is a personal mention @Kamil\\\nThis is a duplicate channel mention @General\\\nThis is a team mention @pzsp2z1teams\\\nYour order #12345 is ready.\\\nThank you!",
		},
		{
			name:  "message with paragraphs and links",
			input: "<p>Check out this link: <a href=\"https://github.com\">GitHub</a></p><p>And this one: <a href=\"https://google.com\">Google</a></p>",
			want:  "Check out this link: [GitHub](https://github.com)\n\nAnd this one: [Google](https://google.com)",
		},
		{
			name:  "message with mentions and formatting",
			input: "<p>Hey <at id=\"0\">Team</at>, please review this <b>important</b> update!</p>",
			want:  "Hey @Team, please review this **important** update!",
		},
		{
			name:  "HTML entities - ampersand",
			input: "Tom &amp; Jerry",
			want:  "Tom & Jerry",
		},
		{
			name:  "HTML entities - less than and greater than",
			input: "5 &lt; 10 &gt; 3",
			want:  "5 < 10 > 3",
		},
		{
			name:  "HTML entities - nbsp",
			input: "Hello&nbsp;World",
			want:  "Hello World",
		},
		{
			name:  "HTML entities - quotes",
			input: "&quot;Hello&quot;",
			want:  "\"Hello\"",
		},
		{
			name:  "whitespace handling",
			input: "  <p>  Some text  </p>  ",
			want:  "Some text",
		},
		{
			name:  "consecutive breaks",
			input: "Line 1<br><br><br>Line 2",
			want:  "Line 1\\\nLine 2",
		},
		{
			name:  "br followed by literal newline",
			input: "Line 1<br>\nLine 2",
			want:  "Line 1\\\n\nLine 2",
		},
		{
			name:  "literal newline followed by br",
			input: "Line 1\n<br>Line 2",
			want:  "Line 1\n\\\nLine 2",
		},
		{
			name:  "multiple br and literal newlines mixed",
			input: "Line 1<br>\n<br>\nLine 2",
			want:  "Line 1\\\n\nLine 2",
		},
		{
			name:  "paragraph with literal newlines",
			input: "<p>Para 1\n\n</p>\n<p>Para 2</p>",
			want:  "Para 1\n\nPara 2",
		},
		{
			name:  "br with newline before paragraph",
			input: "Line 1<br>\n<p>Para 1</p>",
			want:  "Line 1\\\n\nPara 1",
		},
		{
			name:  "empty paragraph",
			input: "<p></p>",
			want:  "",
		},
		{
			name:  "empty tags",
			input: "<b></b><i></i>",
			want:  "",
		},
		{
			name:  "link with line break inside",
			input: "<a href=\"https://example.com\">Click<br>here</a>",
			want:  "[Click\\\nhere](https://example.com)",
		},
		{
			name:  "mention with line break after",
			input: "Hello <at id=\"0\">User</at><br>How are you?",
			want:  "Hello @User\\\nHow are you?",
		},
		{
			name:  "unclosed tag (malformed HTML)",
			input: "<p>Unclosed paragraph",
			want:  "Unclosed paragraph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.Format(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMarkdownFormatter_WriteMessages(t *testing.T) {
	formatter := NewMarkdownFormatter()
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	messages := []*MessageView{
		{
			ID:        "msg1",
			Author:    "Alice",
			Timestamp: timestamp,
			Content:   "<p>Hello <b>World</b></p>",
			Context: []ContextItem{
				{Label: "Team", Value: "Team A"},
				{Label: "Channel", Value: "General"},
			},
		},
		{
			ID:        "msg2",
			Author:    "Bob",
			Timestamp: timestamp.Add(time.Hour),
			Content:   "Hi Alice",
			Context: []ContextItem{
				{Label: "Chat", Value: "OneOnOne"},
			},
		},
	}

	expected := "### From Alice\n\n" +
		"**Team:** Team A\\\n**Channel:** General\\\n**Date:** 01 Jan 24 12:00 UTC\\\n**ID:** msg1\n\n" +
		"Hello **World**\n\n" +
		"---\n\n" +
		"### From Bob\n\n" +
		"**Chat:** OneOnOne\\\n**Date:** 01 Jan 24 13:00 UTC\\\n**ID:** msg2\n\n" +
		"Hi Alice\n\n" +
		"---\n\n"

	var buf bytes.Buffer
	err := formatter.WriteMessages(&buf, messages)
	assert.NoError(t, err)

	// Normalize output for comparison (remove potential carriage returns)
	got := strings.ReplaceAll(buf.String(), "\r\n", "\n")
	assert.Equal(t, expected, got)
}
