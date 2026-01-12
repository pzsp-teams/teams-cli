package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// formKeyMap defines all key bindings for the form interface
type formKeyMap struct {
	Quit        key.Binding
	Submit      key.Binding
	NextField   key.Binding
	PrevField   key.Binding
	NextVariant key.Binding
	PrevVariant key.Binding
	ChoiceLeft  key.Binding
	ChoiceRight key.Binding
}

// newFormKeyMap returns a new formKeyMap with all bindings configured
func newFormKeyMap() formKeyMap {
	return formKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab", "ctrl+down"),
			key.WithHelp("tab/ctrl+↓", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "ctrl+up"),
			key.WithHelp("shift+tab/ctrl+↑", "prev field"),
		),
		NextVariant: key.NewBinding(
			key.WithKeys("ctrl+right", "pgdown", "ctrl+l"),
			key.WithHelp("ctrl+→/pgdn", "next variant"),
		),
		PrevVariant: key.NewBinding(
			key.WithKeys("ctrl+left", "pgup", "ctrl+h"),
			key.WithHelp("ctrl+←/pgup", "prev variant"),
		),
		ChoiceLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev choice"),
		),
		ChoiceRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next choice"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k formKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextField, k.PrevField, k.Submit, k.Quit}
}

// ShortHelpWithVariants returns keybindings including variant navigation
func (k formKeyMap) ShortHelpWithVariants() []key.Binding {
	return []key.Binding{k.NextField, k.PrevField, k.NextVariant, k.PrevVariant, k.Submit, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k formKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextField, k.PrevField, k.ChoiceLeft, k.ChoiceRight},
		{k.Submit, k.Quit},
	}
}

// FullHelpWithVariants returns keybindings including variant navigation
func (k formKeyMap) FullHelpWithVariants() [][]key.Binding {
	return [][]key.Binding{
		{k.NextField, k.PrevField, k.ChoiceLeft, k.ChoiceRight},
		{k.NextVariant, k.PrevVariant, k.Submit, k.Quit},
	}
}

// Helper methods for checking key matches

func (k formKeyMap) isQuit(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.Quit)
}

func (k formKeyMap) isSubmit(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.Submit)
}

func (k formKeyMap) isNextField(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.NextField)
}

func (k formKeyMap) isPrevField(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.PrevField)
}

func (k formKeyMap) isNextVariant(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.NextVariant)
}

func (k formKeyMap) isPrevVariant(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.PrevVariant)
}

func (k formKeyMap) isChoiceLeft(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.ChoiceLeft)
}

func (k formKeyMap) isChoiceRight(msg tea.KeyMsg) bool {
	return key.Matches(msg, k.ChoiceRight)
}

func (k formKeyMap) isNavKey(msg tea.KeyMsg) bool {
	return k.isNextField(msg) || k.isPrevField(msg) || k.isSubmit(msg)
}
