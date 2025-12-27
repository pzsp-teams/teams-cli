package templates

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pzsp-teams/cli/internal/file_readers"
)

func closeFile(t *testing.T, file *os.File) {
	if err := file.Close(); err != nil {
		t.Logf("failed to close file: %v", err)
	}
}

func TestRegistry_ParseJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	content := `{
		"channel1": {
			"name": "Alice",
			"email": "alice@example.com"
		},
		"channel2": {
			"name": "Bob",
			"email": "bob@example.com"
		}
	}`

	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	registry := NewParserRegistry()
	parser, err := registry.GetParser(filePath)
	if err != nil {
		t.Fatalf("Registry.GetParser() unexpected error: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer closeFile(t, file)
	messages := make(map[string]file_readers.TemplateData)
	messagesErr := parser(file, &messages)
	if messagesErr != nil {
		t.Fatalf("Parser.Parse() unexpected error: %v", messagesErr)
	}

	if len(messages) != 2 {
		t.Errorf("Parser.Parse() got %d messages, want 2", len(messages))
	}

	if messages["channel1"]["name"] != "Alice" {
		t.Errorf("Parser.Parse() channel1 name = %q, want %q", messages["channel1"]["name"], "Alice")
	}

	if messages["channel2"]["name"] != "Bob" {
		t.Errorf("Parser.Parse() channel2 name = %q, want %q", messages["channel2"]["name"], "Bob")
	}
}

func TestRegistry_ParseYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yaml")

	content := `channel1:
  name: Alice
  email: alice@example.com
channel2:
  name: Bob
  email: bob@example.com
`

	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	registry := NewParserRegistry()
	parser, err := registry.GetParser(filePath)
	if err != nil {
		t.Fatalf("Registry.GetParser() unexpected error: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer closeFile(t, file)
	messages := make(map[string]file_readers.TemplateData)
	messagesErr := parser(file, messages)
	if messagesErr != nil {
		t.Fatalf("Parser.Parse() unexpected error: %v", messagesErr)
	}

	if len(messages) != 2 {
		t.Errorf("Parser.Parse() got %d messages, want 2", len(messages))
	}

	if messages["channel1"]["name"] != "Alice" {
		t.Errorf("Parser.Parse() channel1 name = %q, want %q", messages["channel1"]["name"], "Alice")
	}

	if messages["channel2"]["name"] != "Bob" {
		t.Errorf("Parser.Parse() channel2 name = %q, want %q", messages["channel2"]["name"], "Bob")
	}
}

func TestRegistry_ParseYMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yml")

	content := `channel1:
  name: Alice
channel2:
  name: Bob
`

	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	registry := NewParserRegistry()
	parser, err := registry.GetParser(filePath)
	if err != nil {
		t.Fatalf("Registry.GetParser() unexpected error: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer closeFile(t, file)

	messages := make(map[string]file_readers.TemplateData)
	messagesErr := parser(file, messages)
	if messagesErr != nil {
		t.Fatalf("Parser.Parse() unexpected error: %v", messagesErr)
	}

	if len(messages) != 2 {
		t.Errorf("Parser.Parse() got %d messages, want 2", len(messages))
	}

	if messages["channel1"]["name"] != "Alice" {
		t.Errorf("Parser.Parse() channel1 name = %q, want %q", messages["channel1"]["name"], "Alice")
	}
}

func TestRegistry_ParseTOMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.toml")

	content := `[channel1]
name = "Alice"
email = "alice@example.com"

[channel2]
name = "Bob"
email = "bob@example.com"
`

	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	registry := NewParserRegistry()
	parser, err := registry.GetParser(filePath)
	if err != nil {
		t.Fatalf("Registry.GetParser() unexpected error: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer closeFile(t, file)

	messages := make(map[string]file_readers.TemplateData)
	messagesErr := parser(file, &messages)
	if messagesErr != nil {
		t.Fatalf("Parser.Parse() unexpected error: %v", messagesErr)
	}

	if len(messages) != 2 {
		t.Errorf("Parser.Parse() got %d messages, want 2", len(messages))
	}

	if messages["channel1"]["name"] != "Alice" {
		t.Errorf("Parser.Parse() channel1 name = %q, want %q", messages["channel1"]["name"], "Alice")
	}

	if messages["channel2"]["name"] != "Bob" {
		t.Errorf("Parser.Parse() channel2 name = %q, want %q", messages["channel2"]["name"], "Bob")
	}
}

func TestRegistry_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xml")

	err := os.WriteFile(filePath, []byte("<data></data>"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	registry := NewParserRegistry()
	_, err = registry.GetParser(filePath)

	if err == nil {
		t.Error("Registry.GetParser() expected error for unsupported format, got nil")
	}
}

func TestRegistry_SupportedFormats(t *testing.T) {
	registry := NewParserRegistry()
	formats := registry.SupportedFormats()

	expectedFormats := map[string]bool{
		"json": true,
		"yaml": true,
		"yml":  true,
		"toml": true,
	}

	if len(formats) != len(expectedFormats) {
		t.Errorf("Registry.SupportedFormats() got %d formats, want %d", len(formats), len(expectedFormats))
	}

	for _, format := range formats {
		if !expectedFormats[format] {
			t.Errorf("Registry.SupportedFormats() unexpected format: %s", format)
		}
	}
}

func TestRegistry_CustomParser(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.custom")

	registry := NewParserRegistry()

	customParser := file_readers.DecodeJSON
	registry.Register("custom", customParser)

	parser, err := registry.GetParser(filePath)
	if err != nil {
		t.Fatalf("Registry.GetParser() unexpected error: %v", err)
	}

	if reflect.ValueOf(parser).Pointer() != reflect.ValueOf(customParser).Pointer() {
		t.Error("Registry.GetParser() did not return registered custom parser")
	}
}
