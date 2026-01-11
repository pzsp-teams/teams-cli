package tui

import (
	"context"
	"testing"

	"github.com/pzsp-teams/cli/app"
)

var mockRegistry = []app.CommandDef{
	{
		Use:   "teams",
		Short: "Teams - Manage teams",
	},
	{
		Use:   "channels",
		Short: "Channels - Manage channels",
	},
	{
		Use:   "chats",
		Short: "Chats - Manage chats",
	},
}

func TestInitialModel(t *testing.T) {
	m := initialModel(context.Background(), mockRegistry, "")

	if m.list.Title != "Teams CLI - Main Menu" {
		t.Errorf("Expected title 'Teams CLI - Main Menu', got '%s'", m.list.Title)
	}

	items := m.list.Items()
	expectedItemCount := 3
	if len(items) != expectedItemCount {
		t.Errorf("Expected %d menu items, got %d", expectedItemCount, len(items))
	}

	expectedItems := []string{
		"Teams - Manage teams",
		"Channels - Manage channels",
		"Chats - Manage chats",
	}

	for i, expected := range expectedItems {
		if i >= len(items) {
			t.Errorf("Missing menu item at index %d: expected '%s'", i, expected)
			continue
		}

		item, ok := items[i].(item)
		if !ok {
			t.Errorf("Item at index %d is not of type 'item'", i)
			continue
		}

		if string(item) != expected {
			t.Errorf("Item at index %d: expected '%s', got '%s'", i, expected, string(item))
		}
	}
}

func TestItemDelegate(t *testing.T) {
	delegate := itemDelegate{}

	if delegate.Height() != 1 {
		t.Errorf("Expected Height() to return 1, got %d", delegate.Height())
	}

	if delegate.Spacing() != 0 {
		t.Errorf("Expected Spacing() to return 0, got %d", delegate.Spacing())
	}
}

func TestModelInit(t *testing.T) {
	m := initialModel(context.Background(), mockRegistry, "")
	cmd := m.Init()

	if cmd == nil {
		t.Errorf("Expected Init() to return a command, got nil")
	}
}

func TestItemFilterValue(t *testing.T) {
	testItem := item("Test Item")

	if testItem.FilterValue() != "Test Item" {
		t.Errorf("Expected FilterValue() to return 'Test Item', got '%s'", testItem.FilterValue())
	}
}