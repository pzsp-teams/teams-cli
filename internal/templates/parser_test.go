package templates

import (
	"strings"
	"testing"

	"github.com/pzsp-teams/cli/internal/file_readers"
	"github.com/pzsp-teams/lib/models"
)

func TestMessageParser_JSONFormatMultipleRecipients(t *testing.T) {
	template := "Hello {{.name}}! Your order #{{.order}} is ready."
	data := `{
		"alice": {
			"name": "Alice",
			"order": "12345"
		},
		"bob": {
			"name": "Bob",
			"order": "67890"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMessages := map[string]string{
		"alice": "Hello Alice! Your order #12345 is ready.",
		"bob":   "Hello Bob! Your order #67890 is ready.",
	}

	if len(messages) != len(wantMessages) {
		t.Errorf("MessageParser.Parse() got %d messages, want %d", len(messages), len(wantMessages))
	}

	for recipient, wantMsg := range wantMessages {
		gotMsg, ok := messages[recipient]
		if !ok {
			t.Errorf("MessageParser.Parse() missing message for recipient %q", recipient)
			continue
		}
		if gotMsg != wantMsg {
			t.Errorf("MessageParser.Parse() for recipient %q:\ngot:\n%s\nwant:\n%s", recipient, gotMsg, wantMsg)
		}
	}
}

func TestMessageParser_TemplateWithMultiplePlaceholders(t *testing.T) {
	template := `Hello {{.name}}!

Meeting reminder:
Date: {{.date}}
Time: {{.time}}
Topic: {{.topic}}

See you there!`
	data := `{
		"team1": {
			"name": "Team Alpha",
			"date": "December 1, 2025",
			"time": "10:00",
			"topic": "Q4 Review"
		},
		"team2": {
			"name": "Team Beta",
			"date": "December 1, 2025",
			"time": "11:00",
			"topic": "Project Review"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMessages := map[string]string{
		"team1": `Hello Team Alpha!<br><br>Meeting reminder:<br>Date: December 1, 2025<br>Time: 10:00<br>Topic: Q4 Review<br><br>See you there!`,
		"team2": `Hello Team Beta!<br><br>Meeting reminder:<br>Date: December 1, 2025<br>Time: 11:00<br>Topic: Project Review<br><br>See you there!`,
	}

	if len(messages) != len(wantMessages) {
		t.Errorf("MessageParser.Parse() got %d messages, want %d", len(messages), len(wantMessages))
	}

	for recipient, wantMsg := range wantMessages {
		gotMsg, ok := messages[recipient]
		if !ok {
			t.Errorf("MessageParser.Parse() missing message for recipient %q", recipient)
			continue
		}
		if gotMsg != wantMsg {
			t.Errorf("MessageParser.Parse() for recipient %q:\ngot:\n%s\nwant:\n%s", recipient, gotMsg, wantMsg)
		}
	}
}

func TestMessageParser_YAMLFormat(t *testing.T) {
	template := "Hello {{.name}}!"
	data := `
alice:
  name: Alice
bob:
  name: Bob
`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeYAML)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMessages := map[string]string{
		"alice": "Hello Alice!",
		"bob":   "Hello Bob!",
	}

	if len(messages) != len(wantMessages) {
		t.Errorf("MessageParser.Parse() got %d messages, want %d", len(messages), len(wantMessages))
	}

	for recipient, wantMsg := range wantMessages {
		gotMsg, ok := messages[recipient]
		if !ok {
			t.Errorf("MessageParser.Parse() missing message for recipient %q", recipient)
			continue
		}
		if gotMsg != wantMsg {
			t.Errorf("MessageParser.Parse() for recipient %q got %q, want %q", recipient, gotMsg, wantMsg)
		}
	}
}

func TestMessageParser_TOMLFormat(t *testing.T) {
	template := "Hello {{.name}}!"
	data := `
[alice]
name = "Alice"

[bob]
name = "Bob"
`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeTOML)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMessages := map[string]string{
		"alice": "Hello Alice!",
		"bob":   "Hello Bob!",
	}

	if len(messages) != len(wantMessages) {
		t.Errorf("MessageParser.Parse() got %d messages, want %d", len(messages), len(wantMessages))
	}

	for recipient, wantMsg := range wantMessages {
		gotMsg, ok := messages[recipient]
		if !ok {
			t.Errorf("MessageParser.Parse() missing message for recipient %q", recipient)
			continue
		}
		if gotMsg != wantMsg {
			t.Errorf("MessageParser.Parse() for recipient %q got %q, want %q", recipient, gotMsg, wantMsg)
		}
	}
}

func TestMessageParser_InvalidTemplateSyntax(t *testing.T) {
	template := "Hello {{.name}"
	data := `{"alice": {"name": "Alice"}}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	_, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)

	if err == nil {
		t.Error("NewMessageParser() expected error for invalid template syntax, got nil")
	}
}

func TestMessageParser_InvalidJSONData(t *testing.T) {
	template := "Hello {{.name}}!"
	data := `{"alice": invalid json}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	_, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)

	if err == nil {
		t.Error("NewMessageParser() expected error for invalid JSON data, got nil")
	}
}

func TestMessageParser_InvalidYAMLData(t *testing.T) {
	template := "Hello {{.name}}!"
	data := `alice:
  name: Alice
    invalid: indentation
  email: test
`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	_, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeYAML)

	if err == nil {
		t.Error("NewMessageParser() expected error for invalid YAML data, got nil")
	}
}

func TestMessageParser_InvalidTOMLData(t *testing.T) {
	template := "Hello {{.name}}!"
	data := `[alice]
name = "Alice"
email = invalid toml without quotes
`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	_, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeTOML)

	if err == nil {
		t.Error("NewMessageParser() expected error for invalid TOML data, got nil")
	}
}

func TestMessageParser_MissingPlaceholderInData(t *testing.T) {
	template := "Hello {{.name}}! Your email is {{.email}}"
	data := `{"alice": {"name": "Alice"}}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	_, err = mp.Parse()

	if err == nil {
		t.Error("MessageParser.Parse() expected error for missing placeholder, got nil")
	}
}

func TestMessageParser_EmptyData(t *testing.T) {
	tmplReader := strings.NewReader("Hello {{.name}}!")
	dataReader := strings.NewReader(`{}`)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Errorf("MessageParser.Parse() unexpected error: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("MessageParser.Parse() got %d messages, want 0", len(messages))
	}
}

func TestMessageParser_PlainTextWithCRLF(t *testing.T) {
	template := "Hello {{.name}}!\r\n\r\nYour order #{{.order}} is ready.\r\nThank you!"
	data := `{
		"alice": {
			"name": "Alice",
			"order": "12345"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "Hello Alice!\r<br>\r<br>Your order #12345 is ready.\r<br>Thank you!"
	gotMsg := messages["alice"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for CRLF:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_HTMLContentPreserved(t *testing.T) {
	template := "<p>Hello {{.name}}!</p>\n\n<p>Your order #{{.order}} is ready.</p>\n<p>Thank you!</p>"
	data := `{
		"alice": {
			"name": "Alice",
			"order": "12345"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "<p>Hello Alice!</p>\n\n<p>Your order #12345 is ready.</p>\n<p>Thank you!</p>"
	gotMsg := messages["alice"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for HTML content:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_HTMLContentWithCRLF(t *testing.T) {
	template := "<p>Hello {{.name}}!</p>\r\n\r\n<p>Your order #{{.order}} is ready.</p>\r\n<p>Thank you!</p>"
	data := `{
		"bob": {
			"name": "Bob",
			"order": "67890"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "<p>Hello Bob!</p>\r\n\r\n<p>Your order #67890 is ready.</p>\r\n<p>Thank you!</p>"
	gotMsg := messages["bob"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for HTML with CRLF:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_HTMLWithBoldAndItalic(t *testing.T) {
	template := "Hello <b>{{.name}}</b>!\n\nYour <i>special</i> order is ready."
	data := `{
		"charlie": {
			"name": "Charlie"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "Hello <b>Charlie</b>!\n\nYour <i>special</i> order is ready."
	gotMsg := messages["charlie"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for HTML with bold/italic:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_HTMLWithLinks(t *testing.T) {
	template := "Hello {{.name}}!\n\nClick <a href=\"https://example.com\">here</a> for details."
	data := `{
		"dave": {
			"name": "Dave"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "Hello Dave!\n\nClick <a href=\"https://example.com\">here</a> for details."
	gotMsg := messages["dave"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for HTML with links:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_HTMLWithBRTag(t *testing.T) {
	template := "Hello {{.name}}!<br><br>Your order is ready."
	data := `{
		"eve": {
			"name": "Eve"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "Hello Eve!<br><br>Your order is ready."
	gotMsg := messages["eve"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for HTML with <br>:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestMessageParser_MixedCRLFAndLF(t *testing.T) {
	template := "Line 1\nLine 2\r\nLine 3\nLine 4"
	data := `{
		"frank": {
			"name": "Frank"
		}
	}`

	tmplReader := strings.NewReader(template)
	dataReader := strings.NewReader(data)

	mp, err := NewMessageParser(tmplReader, dataReader, file_readers.DecodeJSON)
	if err != nil {
		t.Fatalf("NewMessageParser() unexpected error: %v", err)
	}
	messages, err := mp.Parse()
	if err != nil {
		t.Fatalf("MessageParser.Parse() unexpected error: %v", err)
	}

	wantMsg := "Line 1<br>Line 2\r<br>Line 3<br>Line 4"
	gotMsg := messages["frank"]

	if gotMsg != wantMsg {
		t.Errorf("MessageParser.Parse() for mixed CRLF and LF:\ngot:\n%q\nwant:\n%q", gotMsg, wantMsg)
	}
}

func TestExtractMentions_NoMentions(t *testing.T) {
	content := "Hello world! This is a regular message."
	mentions := ExtractMentions(content)

	if mentions != nil {
		t.Errorf("ExtractMentions() got %v, want nil", mentions)
	}
}

func TestExtractMentions_SingleMention(t *testing.T) {
	content := "Hello @@alice@@ how are you?"
	mentions := ExtractMentions(content)

	want := []string{"alice"}
	if len(mentions) != len(want) {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
		return
	}
	if mentions[0] != want[0] {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
	}
}

func TestExtractMentions_MultipleMentions(t *testing.T) {
	content := "Hello @@alice@@ and @@bob@@, welcome!"
	mentions := ExtractMentions(content)

	want := []string{"alice", "bob"}
	if len(mentions) != len(want) {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
		return
	}
	for i, m := range mentions {
		if m != want[i] {
			t.Errorf("ExtractMentions()[%d] got %v, want %v", i, m, want[i])
		}
	}
}

func TestExtractMentions_DuplicateMentions(t *testing.T) {
	content := "Hello @@alice@@, this is for @@alice@@ again"
	mentions := ExtractMentions(content)

	want := []string{"alice", "alice"}
	if len(mentions) != len(want) {
		t.Errorf("ExtractMentions() got %v mentions, want %v (duplicates should be included)", len(mentions), len(want))
		return
	}
	for i, m := range mentions {
		if m != want[i] {
			t.Errorf("ExtractMentions()[%d] got %v, want %v", i, m, want[i])
		}
	}
}

func TestExtractMentions_WithWhitespace(t *testing.T) {
	content := "Hello @@   alice  @@ how are you?"
	mentions := ExtractMentions(content)

	want := []string{"alice"}
	if len(mentions) != len(want) {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
		return
	}
	if mentions[0] != want[0] {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
	}
}

func TestExtractMentions_EmailFormat(t *testing.T) {
	content := "Please contact @@alice@example.com@@ for details"
	mentions := ExtractMentions(content)

	want := []string{"alice@example.com"}
	if len(mentions) != len(want) {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
		return
	}
	if mentions[0] != want[0] {
		t.Errorf("ExtractMentions() got %v, want %v", mentions, want)
	}
}

func TestReplaceMentions_NoMentions(t *testing.T) {
	content := "Hello world!"
	mentions := []models.Mention{}
	result := ReplaceMentions(content, mentions)

	if result != content {
		t.Errorf("ReplaceMentions() got %q, want %q", result, content)
	}
}

func TestReplaceMentions_NilMentions(t *testing.T) {
	content := "Hello world!"
	result := ReplaceMentions(content, nil)

	if result != content {
		t.Errorf("ReplaceMentions() got %q, want %q", result, content)
	}
}

func TestReplaceMentions_SingleMention(t *testing.T) {
	content := "Hello @@ @alice@@!"
	mentions := []models.Mention{
		{
			Kind:     models.MentionUser,
			AtID:     0,
			Text:     "Alice Smith",
			TargetID: "user123",
		},
	}

	result := ReplaceMentions(content, mentions)
	want := `Hello <at id="0">Alice Smith</at>!`

	if result != want {
		t.Errorf("ReplaceMentions() got %q, want %q", result, want)
	}
}

func TestReplaceMentions_MultipleMentions(t *testing.T) {
	content := "Hello @@alice@@ and @@bob@@!"
	mentions := []models.Mention{
		{
			Kind:     models.MentionUser,
			AtID:     0,
			Text:     "Alice Smith",
			TargetID: "user123",
		},
		{
			Kind:     models.MentionUser,
			AtID:     1,
			Text:     "Bob Jones",
			TargetID: "user456",
		},
	}

	result := ReplaceMentions(content, mentions)
	want := `Hello <at id="0">Alice Smith</at> and <at id="1">Bob Jones</at>!`

	if result != want {
		t.Errorf("ReplaceMentions() got %q, want %q", result, want)
	}
}

func TestReplaceMentions_WithWhitespace(t *testing.T) {
	content := "Hello @@    alice  @@!"
	mentions := []models.Mention{
		{
			Kind:     models.MentionUser,
			AtID:     0,
			Text:     "Alice Smith",
			TargetID: "user123",
		},
	}

	result := ReplaceMentions(content, mentions)
	want := `Hello <at id="0">Alice Smith</at>!`

	if result != want {
		t.Errorf("ReplaceMentions() got %q, want %q", result, want)
	}
}

func TestReplaceMentions_LessMentionsThanPatterns(t *testing.T) {
	content := "Hello @@alice@@ and @@bob@@!"
	mentions := []models.Mention{
		{
			Kind:     models.MentionUser,
			AtID:     0,
			Text:     "Alice Smith",
			TargetID: "user123",
		},
	}

	result := ReplaceMentions(content, mentions)
	want := `Hello <at id="0">Alice Smith</at> and @@bob@@!`

	if result != want {
		t.Errorf("ReplaceMentions() got %q, want %q", result, want)
	}
}
