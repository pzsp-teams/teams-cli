package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pzsp-teams/cli/app"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = blurredStyle.Render("[ Submit ]")
)

// formField tracks a field in the form and its type
type formField struct {
	fieldType string // "text" or "choice"
	index     int    // Index in inputs or choices array
	flagIndex int    // Index in the CommandDef.Flags array
}

// choiceField represents a radio button group for InputChoice
type choiceField struct {
	flagDef  *app.FlagDef
	selected int
}

// formModel represents a form with mixed input types
type formModel struct {
	def           *app.CommandDef
	inputs        []textinput.Model // Text inputs
	choices       []choiceField     // Radio button groups
	fieldOrder    []formField       // Tracks rendering order
	focused       int
	err           error
	validationErr string
}

func createChoiceField(flagDef *app.FlagDef) choiceField {
	choice := choiceField{
		flagDef:  flagDef,
		selected: 0,
	}

	if flagDef.DefaultVal != nil {
		if defStr, ok := flagDef.DefaultVal.(string); ok {
			for i := range flagDef.Options {
				if flagDef.Options[i] == defStr {
					choice.selected = i
					break
				}
			}
		}
	}

	return choice
}

func createTextInput(flagDef *app.FlagDef, shouldFocus bool) textinput.Model {
	t := textinput.New()
	t.CharLimit = 256
	t.Width = 50
	t.Prompt = flagDef.Name + ": "
	t.Placeholder = flagDef.Usage

	switch flagDef.Type {
	case app.InputFile:
		t.Placeholder += " (File Picker coming soon)"
	case app.InputDate:
		t.Placeholder += " (Date Picker coming soon)"
	case app.InputPassword:
		t.EchoMode = textinput.EchoPassword
	case app.InputBool:
		t.Placeholder += " (true/false)"
	}

	if shouldFocus {
		t.Focus()
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle
	}

	return t
}

// NewFormModel creates a new form model from a command definition
func NewFormModel(def *app.CommandDef) *formModel {
	var inputs []textinput.Model
	var choices []choiceField
	var fieldOrder []formField

	for flagIdx := range def.Flags {
		flagDef := &def.Flags[flagIdx]

		switch flagDef.Type {
		case app.InputChoice:
			choice := createChoiceField(flagDef)
			choices = append(choices, choice)
			fieldOrder = append(fieldOrder, formField{
				fieldType: "choice",
				index:     len(choices) - 1,
				flagIndex: flagIdx,
			})

		default:
			shouldFocus := len(inputs) == 0 && len(choices) == 0
			input := createTextInput(flagDef, shouldFocus)
			inputs = append(inputs, input)
			fieldOrder = append(fieldOrder, formField{
				fieldType: "text",
				index:     len(inputs) - 1,
				flagIndex: flagIdx,
			})
		}
	}

	return &formModel{
		def:        def,
		inputs:     inputs,
		choices:    choices,
		fieldOrder: fieldOrder,
		focused:    0,
	}
}

// Init implements tea.Model
func (m *formModel) Init() tea.Cmd {
	return textinput.Blink
}

var quitKeys = map[string]struct{}{
	"ctrl+c": {},
	"esc":    {},
}

var choiceKeys = map[string]struct{}{
	"left":  {},
	"right": {},
	"h":     {},
	"l":     {},
}

var navKeys = map[string]struct{}{
	"tab": {}, "shift+tab": {}, "enter": {},
	"up": {}, "down": {},
	"ctrl+k": {}, "ctrl+j": {},
}

// Update implements tea.Model
func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	s := key.String()

	if _, found := quitKeys[s]; found {
		return m, tea.Quit
	}

	if _, found := choiceKeys[s]; found {
		if m.handleChoiceSelection(s) {
			return m, nil
		}
	}

	if _, found := navKeys[s]; found {
		return m.handleNavigation(s)
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *formModel) handleChoiceSelection(msg string) bool {
	if m.focused >= len(m.fieldOrder) {
		return false
	}

	field := m.fieldOrder[m.focused]
	if field.fieldType != "choice" {
		return false
	}

	choice := &m.choices[field.index]
	if msg == "left" {
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

func (m *formModel) handleNavigation(msg string) (tea.Model, tea.Cmd) {
	if msg == "enter" && m.focused == len(m.fieldOrder) {
		return m.submitForm()
	}

	m.moveFocus(msg)
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

func (m *formModel) moveFocus(msg string) {
	if msg == "up" || msg == "shift+tab" {
		m.focused--
	} else {
		m.focused++
	}

	if m.focused > len(m.fieldOrder) {
		m.focused = 0
	} else if m.focused < 0 {
		m.focused = len(m.fieldOrder)
	}
}

func (m *formModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i < len(m.inputs); i++ {
		isFocused := false
		if m.focused < len(m.fieldOrder) {
			field := m.fieldOrder[m.focused]
			isFocused = field.fieldType == "text" && field.index == i
		}

		if isFocused {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = noStyle
			m.inputs[i].TextStyle = noStyle
		}
	}

	return tea.Batch(cmds...)
}

func (m *formModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// validateForm validates the form using schema validation
func (m *formModel) validateForm() error {
	flagMap := make(map[string]any)

	for _, field := range m.fieldOrder {
		flagDef := &m.def.Flags[field.flagIndex]

		if field.fieldType == "text" {
			val := m.inputs[field.index].Value()
			if val != "" {
				flagMap[flagDef.Name] = val
			}
		} else { // choice
			choice := m.choices[field.index]
			selectedValue := choice.flagDef.Options[choice.selected]
			flagMap[flagDef.Name] = selectedValue
		}
	}

	return app.ValidateFlags(flagMap, m.def.Flags)
}

func (m *formModel) renderChoice(choice choiceField, focused bool) string {
	var b strings.Builder

	labelStyle := noStyle
	if focused {
		labelStyle = focusedStyle
	}

	b.WriteString(labelStyle.Render(choice.flagDef.Name + ": "))

	for i, opt := range choice.flagDef.Options {
		if i > 0 {
			b.WriteString("  ")
		}

		if i == choice.selected {
			if focused {
				b.WriteString(focusedStyle.Render("(*) " + opt))
			} else {
				b.WriteString(noStyle.Render("(*) " + opt))
			}
		} else {
			b.WriteString(blurredStyle.Render("( ) " + opt))
		}
	}

	return b.String()
}

// View implements tea.Model
func (m *formModel) View() string {
	var b strings.Builder

	for i, field := range m.fieldOrder {
		if field.fieldType == "text" {
			b.WriteString(m.inputs[field.index].View())
		} else { // choice
			choice := m.choices[field.index]
			b.WriteString(m.renderChoice(choice, i == m.focused))
		}
		b.WriteString("\n")
	}

	if m.validationErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.validationErr))
		b.WriteString("\n")
	}

	button := blurredButton
	if m.focused == len(m.fieldOrder) {
		button = focusedButton
	}
	b.WriteString("\n" + button + "\n")

	return b.String()
}
