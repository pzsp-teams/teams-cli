package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pzsp-teams/cli/app"
)

var (
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	noStyle       = lipgloss.NewStyle()
	formHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(1, 0, 0, 2)

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = blurredStyle.Render("[ Submit ]")
)

type variantDirection int

const (
	variantNext variantDirection = iota
	variantPrev
)

// formModel represents a form with mixed input types and variants (mutually exclusive groups)
type formModel struct {
	def       *app.CommandDef
	inputs    []textinput.Model
	choices   []choiceField
	textAreas []textarea.Model

	variants     [][]formField // Groups of fields to show together
	variantIndex int

	focused       int // Index within the current variant
	err           error
	validationErr string

	width  int
	height int

	keys formKeyMap
	help help.Model
}

// NewFormModel creates a new form model from a command definition
func NewFormModel(def *app.CommandDef) *formModel {
	inputs, choices, textAreas, fieldMap := buildFields(def)
	variants := buildVariants(def, fieldMap)

	m := &formModel{
		def:          def,
		inputs:       inputs,
		choices:      choices,
		textAreas:    textAreas,
		variants:     variants,
		variantIndex: 0,
		focused:      0,
		keys:         newFormKeyMap(),
		help:         help.New(),
	}

	m.updateFocus()
	return m
}

// Init implements tea.Model
func (m *formModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink)
}

// Update implements tea.Model
func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	if m.keys.isQuit(key) {
		return m, tea.Quit
	}

	// Handle Variant Navigation
	if m.keys.isNextVariant(key) {
		m.changeVariant(variantNext)
		return m, nil
	} else if m.keys.isPrevVariant(key) {
		m.changeVariant(variantPrev)
		return m, nil
	}

	// If inside a textarea and it's focused, we might want to allow normal navigation keys
	// UNLESS it's a specific key we want to hijack (like Tab to move focus).
	// But `textarea` usually consumes Enter for newlines.
	if m.isFocusedTextArea() {
		// Pass everything to textarea except navigation out keys
		if m.keys.isNextField(key) || m.keys.isPrevField(key) {
			return m.handleNavigation(key)
		}
		// Allow "esc" to blur/navigate? Handled in global key handler in model.go usually.
		// If we want "Enter" to insert newline, we just update inputs.
		// If we want "Enter" to submit form...
		// Usually in multiline text, Enter is newline. Ctrl+Enter or Tab to submit/move.
		// Let's assume standard behavior: Enter = newline in textarea.
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	if m.keys.isChoiceLeft(key) || m.keys.isChoiceRight(key) {
		if m.handleChoiceSelection(key) {
			return m, nil
		}
	}

	if m.keys.isNavKey(key) {
		return m.handleNavigation(key)
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *formModel) changeVariant(dir variantDirection) {
	if dir == variantNext {
		m.variantIndex++
		if m.variantIndex >= len(m.variants) {
			m.variantIndex = 0
		}
	} else {
		m.variantIndex--
		if m.variantIndex < 0 {
			m.variantIndex = len(m.variants) - 1
		}
	}
	// Reset focus to top when changing variants
	m.focused = 0
	m.updateFocus()
}

func (m *formModel) isFocusedTextArea() bool {
	currentVariant := m.variants[m.variantIndex]
	if m.focused >= len(currentVariant) {
		return false
	}
	field := currentVariant[m.focused]
	return field.fieldType == app.InputLongString || field.fieldType == app.InputList
}

func (m *formModel) handleChoiceSelection(key tea.KeyMsg) bool {
	currentVariant := m.variants[m.variantIndex]
	if m.focused >= len(currentVariant) {
		return false
	}

	field := currentVariant[m.focused]
	if field.fieldType != app.InputChoice {
		return false
	}

	choice := &m.choices[field.index]
	if m.keys.isChoiceLeft(key) {
		choice.selected--
		if choice.selected < 0 {
			choice.selected = len(choice.flagDef.Options) - 1
		}
	} else {
		choice.selected++
		if choice.selected >= len(choice.flagDef.Options) {
			choice.selected = 0
		}
	}
	return true
}

func (m *formModel) handleNavigation(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentVariant := m.variants[m.variantIndex]

	if m.keys.isSubmit(key) && m.focused == len(currentVariant) {
		return m.submitForm()
	}

	m.moveFocus(key)
	cmd := m.updateFocus()
	return m, cmd
}

func (m *formModel) submitForm() (tea.Model, tea.Cmd) {
	if err := m.validateForm(); err != nil {
		m.validationErr = err.Error()
		return m, nil
	}
	m.validationErr = ""
	return m, tea.Quit
}

func (m *formModel) moveFocus(key tea.KeyMsg) {
	currentVariant := m.variants[m.variantIndex]

	if m.keys.isPrevField(key) {
		m.focused--
	} else {
		m.focused++
	}

	if m.focused > len(currentVariant) {
		m.focused = 0
	} else if m.focused < 0 {
		m.focused = len(currentVariant)
	}
}

func (m *formModel) updateFocus() tea.Cmd {
	currentVariant := m.variants[m.variantIndex]
	var cmds []tea.Cmd

	// Update Text Inputs
	for i := range m.inputs {
		isFocused := false
		if m.focused < len(currentVariant) {
			field := currentVariant[m.focused]
			isFocused = isTextInput(field.fieldType) && field.index == i
		}

		if isFocused {
			cmds = append(cmds, m.inputs[i].Focus())
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = noStyle
			m.inputs[i].TextStyle = noStyle
		}
	}

	// Update Text Areas
	for i := range m.textAreas {
		isFocused := false
		if m.focused < len(currentVariant) {
			field := currentVariant[m.focused]
			isFocused = (field.fieldType == app.InputLongString || field.fieldType == app.InputList) && field.index == i
		}

		if isFocused {
			cmds = append(cmds, m.textAreas[i].Focus())
			// Textarea doesn't have PromptStyle/TextStyle public fields in the same way,
			// but Focus handles the cursor.
		} else {
			m.textAreas[i].Blur()
		}
	}

	return tea.Batch(cmds...)
}

func (m *formModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	for i := range m.textAreas {
		var cmd tea.Cmd
		m.textAreas[i], cmd = m.textAreas[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// validateForm validates the form using schema validation for CURRENT VARIANT
func (m *formModel) validateForm() error {
	flagMap := make(map[string]any)
	currentVariant := m.variants[m.variantIndex]

	var variantFlags []app.FlagDef

	for _, field := range currentVariant {
		flagDef := &m.def.Flags[field.flagIndex]
		variantFlags = append(variantFlags, *flagDef)

		switch field.fieldType {
		case app.InputLongString:
			val := m.textAreas[field.index].Value()
			if val != "" {
				flagMap[flagDef.Name] = val
			}
		case app.InputList:
			val := m.textAreas[field.index].Value()
			if val != "" {
				items := strings.Split(val, "\n")
				var filtered []string
				for _, item := range items {
					trimmed := strings.TrimSpace(item)
					if trimmed != "" {
						filtered = append(filtered, trimmed)
					}
				}
				if len(filtered) > 0 {
					flagMap[flagDef.Name] = filtered
				}
			}
		case app.InputChoice:
			choice := m.choices[field.index]
			selectedValue := choice.flagDef.Options[choice.selected]
			flagMap[flagDef.Name] = selectedValue
		default: // All other text inputs
			val := m.inputs[field.index].Value()
			if val != "" {
				flagMap[flagDef.Name] = val
			}
		}
	}

	return app.ValidateFlags(flagMap, variantFlags)
}

// View implements tea.Model
func (m *formModel) View() string {
	var b strings.Builder

	// Variant header
	if len(m.variants) > 1 {
		fmt.Fprintf(&b, "Variant %d of %d\n", m.variantIndex+1, len(m.variants))
	}

	currentVariant := m.variants[m.variantIndex]

	// Render form fields
	for i, field := range currentVariant {
		isFocused := (i == m.focused)

		switch field.fieldType {
		case app.InputLongString, app.InputList:
			flagName := m.def.Flags[field.flagIndex].Name
			b.WriteString(m.renderTextArea(&m.textAreas[field.index], flagName, isFocused))
		case app.InputChoice:
			choice := m.choices[field.index]
			b.WriteString(m.renderChoice(choice, isFocused))
		default: // Text inputs
			b.WriteString(m.inputs[field.index].View())
		}
		b.WriteString("\n")
		// Extra spacing for text areas if focused
		if (field.fieldType == app.InputLongString || field.fieldType == app.InputList) && isFocused {
			b.WriteString("\n")
		}
	}

	// Validation errors
	if m.validationErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.validationErr))
		b.WriteString("\n")
	}

	// Submit button
	button := blurredButton
	if m.focused == len(currentVariant) {
		button = focusedButton
	}
	b.WriteString("\n" + button + "\n")

	formContent := b.String()

	// Render help with appropriate keybindings
	var helpView string
	if len(m.variants) > 1 {
		// Create a temporary key map that includes variant navigation
		helpView = m.help.ShortHelpView(m.keys.ShortHelpWithVariants())
	} else {
		helpView = m.help.View(m.keys)
	}

	// Calculate lines used by form content
	contentLines := strings.Count(formContent, "\n")
	helpLines := strings.Count(helpView, "\n") + 1 // +1 for the help itself

	// Calculate padding needed to push help to bottom
	// Account for top margin (1 line from model.go)
	usedLines := 1 + contentLines + helpLines
	paddingLines := 0
	if m.height > usedLines {
		paddingLines = m.height - usedLines
	}

	// Build final output with help at bottom
	var final strings.Builder
	final.WriteString(formContent)
	final.WriteString(strings.Repeat("\n", paddingLines))
	final.WriteString(formHelpStyle.Render(helpView))

	return final.String()
}
