package tui

import (
	"testing"

	"github.com/pzsp-teams/cli/app"
)

func TestCalculateVariants(t *testing.T) {
	tests := []struct {
		name     string
		flags    []app.FlagDef
		expected int // Number of variants expected
	}{
		{
			name: "No conflicts",
			flags: []app.FlagDef{
				{Name: "a"},
				{Name: "b"},
			},
			expected: 1,
		},
		{
			name: "Simple conflict",
			flags: []app.FlagDef{
				{Name: "a", ConflictsWith: []string{"b"}},
				{Name: "b", ConflictsWith: []string{"a"}},
			},
			expected: 2, // {a}, {b}
		},
		{
			name: "Complex conflicts",
			flags: []app.FlagDef{
				{Name: "a", ConflictsWith: []string{"b"}},
				{Name: "b", ConflictsWith: []string{"a", "c"}},
				{Name: "c", ConflictsWith: []string{"b"}},
			},
			expected: 2, // {a, c}, {b}
		},
		{
			name: "Disconnected conflicts",
			flags: []app.FlagDef{
				{Name: "a", ConflictsWith: []string{"b"}},
				{Name: "b", ConflictsWith: []string{"a"}},
				{Name: "c", ConflictsWith: []string{"d"}},
				{Name: "d", ConflictsWith: []string{"c"}},
			},
			expected: 4, // {a, c}, {a, d}, {b, c}, {b, d} are all maximal independent sets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variants := calculateVariants(tt.flags)
			if len(variants) != tt.expected {
				t.Errorf("expected %d variants, got %d", tt.expected, len(variants))
			}
		})
	}
}

func TestFormModelNavigation(t *testing.T) {
	def := &app.CommandDef{
		Flags: []app.FlagDef{
			{Name: "a", ConflictsWith: []string{"b"}},
			{Name: "b", ConflictsWith: []string{"a"}},
		},
	}

	m := NewFormModel(def)

	if len(m.variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(m.variants))
	}

	if m.variantIndex != 0 {
		t.Errorf("expected initial variantIndex 0, got %d", m.variantIndex)
	}

	m.changeVariant(variantNext)
	if m.variantIndex != 1 {
		t.Errorf("expected variantIndex 1 after pgdown, got %d", m.variantIndex)
	}

	m.changeVariant(variantNext)
	if m.variantIndex != 0 {
		t.Errorf("expected variantIndex 0 after wrap around, got %d", m.variantIndex)
	}

	m.changeVariant(variantPrev)
	if m.variantIndex != 1 {
		t.Errorf("expected variantIndex 1 after pgup wrap around, got %d", m.variantIndex)
	}
}

func TestIsTextInput(t *testing.T) {
	if !isTextInput(app.InputString) {
		t.Error("expected InputString to be text input")
	}
	if !isTextInput(app.InputInt) {
		t.Error("expected InputInt to be text input")
	}
	if isTextInput(app.InputChoice) {
		t.Error("expected InputChoice NOT to be text input")
	}
	if isTextInput(app.InputLongString) {
		t.Error("expected InputLongString NOT to be text input")
	}
	if isTextInput(app.InputNone) {
		t.Error("expected InputNone NOT to be text input")
	}
}
