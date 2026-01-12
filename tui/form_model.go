package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bubbledatetimepicker "github.com/lcc/bubble-datetime-picker"

	"github.com/pzsp-teams/teams-cli/app"
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
	def         *app.CommandDef
	inputs      []textinput.Model
	choices     []choiceField
	checkboxes  []checkboxField
	textAreas   []textarea.Model
	filePickers []filepicker.Model
	datePickers []*bubbledatetimepicker.DateAndHourModel

	variants     [][]formField // Groups of fields to show together
	variantIndex int

	focused       int // Index within the current variant
	validationErr string

	width  int
	height int

	keys formKeyMap
	help help.Model
}

// NewFormModel creates a new form model from a command definition
func NewFormModel(def *app.CommandDef) *formModel {
	components := buildFields(def)
	variants := buildVariants(def, components.fieldMap)

	h := help.New()
	h.Width = 80

	m := &formModel{
		def:          def,
		inputs:       components.inputs,
		choices:      components.choices,
		checkboxes:   components.checkboxes,
		textAreas:    components.textAreas,
		filePickers:  components.filePickers,
		datePickers:  components.datePickers,
		variants:     variants,
		variantIndex: 0,
		focused:      0,
		keys:         newFormKeyMap(),
		help:         h,
	}

	m.updateFocus()
	return m
}

// Init implements tea.Model
func (m *formModel) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink, textarea.Blink}
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = max(1, msg.Width-2)
		return m, nil
	case datePickerDoneMsg:
		m.focused++
		currentVariant := m.variants[m.variantIndex]
		if m.focused > len(currentVariant) {
			m.focused = 0
		}
		cmd := m.updateFocus()
		return m, cmd
	}

	key, ok := msg.(tea.KeyMsg)
	var cmd tea.Cmd
	if !ok {
		cmd = m.updateInputs(msg)
		return m, cmd
	}

	if m.keys.isQuit(key) {
		return m, tea.Quit
	}

	if handled := m.handleVariantKey(key); handled {
		return m, nil
	}

	if m.isFocusedSpecialField() {
		return m.handleSpecialFieldUpdate(msg, key)
	}

	if m.handleChoiceFieldUpdate(key) {
		return m, nil
	}

	if m.keys.isNavKey(key) {
		return m.handleNavigation(key)
	}

	cmd = m.updateInputs(msg)
	return m, cmd
}

func (m *formModel) handleVariantKey(key tea.KeyMsg) bool {
	if m.keys.isNextVariant(key) {
		m.changeVariant(variantNext)
		return true
	} else if m.keys.isPrevVariant(key) {
		m.changeVariant(variantPrev)
		return true
	}
	return false
}

func (m *formModel) handleSpecialFieldUpdate(msg tea.Msg, key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.keys.isNextField(key) || m.keys.isPrevField(key) {
		return m.handleNavigation(key)
	}

	// For textareas: Enter is newline
	// For file pickers: Enter selects file/enters directory, arrows navigate
	// For date pickers: Enter moves from date to time (handled by picker internally)
	// All handle their own input through updateInputs
	cmd := m.updateInputs(msg)

	// Check for FilePicker auto-advance
	currentVariant := m.variants[m.variantIndex]
	if m.focused < len(currentVariant) {
		field := currentVariant[m.focused]
		if field.fieldType == app.InputFile && m.keys.isSubmit(key) {
			if m.filePickers[field.index].Path != "" {
				return m.handleNavigation(key)
			}
		}
	}

	return m, cmd
}

func (m *formModel) handleChoiceFieldUpdate(key tea.KeyMsg) bool {
	currentVariant := m.variants[m.variantIndex]
	if m.focused >= len(currentVariant) {
		return false
	}
	field := currentVariant[m.focused]

	if field.fieldType == app.InputChoice {
		if m.keys.isChoiceLeft(key) || m.keys.isChoiceRight(key) {
			return m.handleChoiceSelection(key)
		}
	}

	if field.fieldType == app.InputBool {
		if m.keys.isSubmit(key) {
			return m.handleCheckboxToggle()
		}
	}

	return false
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
	m.focused = 0
	m.updateFocus()
}

func (m *formModel) isFocusedSpecialField() bool {
	currentVariant := m.variants[m.variantIndex]
	if m.focused >= len(currentVariant) {
		return false
	}
	field := currentVariant[m.focused]
	return field.fieldType == app.InputLongString ||
		field.fieldType == app.InputList ||
		field.fieldType == app.InputFile ||
		field.fieldType == app.InputDate
}

func (m *formModel) handleChoiceSelection(key tea.KeyMsg) bool {
	currentVariant := m.variants[m.variantIndex]
	field := currentVariant[m.focused]

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

func (m *formModel) handleCheckboxToggle() bool {
	currentVariant := m.variants[m.variantIndex]
	field := currentVariant[m.focused]

	checkbox := &m.checkboxes[field.index]
	checkbox.checked = !checkbox.checked
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

	for i := range m.textAreas {
		isFocused := false
		if m.focused < len(currentVariant) {
			field := currentVariant[m.focused]
			isFocused = (field.fieldType == app.InputLongString || field.fieldType == app.InputList) && field.index == i
		}

		if isFocused {
			cmds = append(cmds, m.textAreas[i].Focus())
		} else {
			m.textAreas[i].Blur()
		}
	}

	return tea.Batch(cmds...)
}

func (m *formModel) updateInputs(msg tea.Msg) tea.Cmd {
	currentVariant := m.variants[m.variantIndex]
	if m.focused >= len(currentVariant) {
		return nil
	}

	field := currentVariant[m.focused]

	switch field.fieldType {
	case app.InputFile:
		var cmd tea.Cmd
		m.filePickers[field.index], cmd = m.filePickers[field.index].Update(msg)
		return cmd

	case app.InputDate:
		updated, updateCmd := m.datePickers[field.index].Update(msg)
		if dp, ok := updated.(*bubbledatetimepicker.DateAndHourModel); ok {
			m.datePickers[field.index] = dp
		}
		return wrapDatePickerCmd(updateCmd)

	case app.InputLongString, app.InputList:
		var cmd tea.Cmd
		m.textAreas[field.index], cmd = m.textAreas[field.index].Update(msg)
		return cmd

	case app.InputChoice, app.InputBool:
		return nil

	default:
		// Regular text input
		if isTextInput(field.fieldType) {
			var cmd tea.Cmd
			m.inputs[field.index], cmd = m.inputs[field.index].Update(msg)
			return cmd
		}
	}

	return nil
}

// collectFlags collects values from all fields in the current variant
func (m *formModel) collectFlags() map[string]any {
	flagMap := make(map[string]any)
	currentVariant := m.variants[m.variantIndex]

	for _, field := range currentVariant {
		flagDef := &m.def.Flags[field.flagIndex]

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
		case app.InputBool:
			checkbox := m.checkboxes[field.index]
			flagMap[flagDef.Name] = checkbox.checked
		case app.InputFile:
			selectedFile := m.filePickers[field.index].Path
			if selectedFile != "" {
				flagMap[flagDef.Name] = selectedFile
			}
		case app.InputDate:
			picker := m.datePickers[field.index]
			dateTime := picker.Time()
			flagMap[flagDef.Name] = dateTime.Format("2006-01-02T15:04:05Z07:00")
		default: // All other text inputs
			val := m.inputs[field.index].Value()
			if val != "" {
				flagMap[flagDef.Name] = val
			}
		}
	}
	return flagMap
}

// validateForm validates the form using schema validation for current variant
func (m *formModel) validateForm() error {
	flagMap := m.collectFlags()
	currentVariant := m.variants[m.variantIndex]

	var variantFlags []app.FlagDef
	for _, field := range currentVariant {
		flagDef := &m.def.Flags[field.flagIndex]
		variantFlags = append(variantFlags, *flagDef)
	}

	return app.ValidateFlags(flagMap, variantFlags)
}

// View implements tea.Model
func (m *formModel) View() string {
	var b strings.Builder

	m.renderVariantHeader(&b)
	m.renderFields(&b)
	m.renderValidationErrors(&b)
	m.renderSubmitButton(&b)

	formContent := b.String()
	helpView := m.renderHelp()

	contentLines := strings.Count(formContent, "\n")
	helpLines := strings.Count(helpView, "\n") + 1 // +1 for the help content itself
	styleExtraLines := 1

	usedLines := 1 + contentLines + helpLines + styleExtraLines
	paddingLines := 0
	if m.height > usedLines {
		paddingLines = m.height - usedLines
	}

	return formContent + strings.Repeat("\n", paddingLines) + formHelpStyle.Render(helpView)
}

func (m *formModel) renderVariantHeader(b *strings.Builder) {
	if len(m.variants) > 1 {
		_, _ = fmt.Fprintf(b, "Variant %d of %d\n", m.variantIndex+1, len(m.variants))
	}
}

func (m *formModel) renderFields(b *strings.Builder) {
	currentVariant := m.variants[m.variantIndex]

	for i, field := range currentVariant {
		isFocused := (i == m.focused)

		switch field.fieldType {
		case app.InputLongString, app.InputList:
			flagName := m.def.Flags[field.flagIndex].Name
			b.WriteString(m.renderTextArea(&m.textAreas[field.index], flagName, isFocused))
		case app.InputChoice:
			choice := m.choices[field.index]
			b.WriteString(m.renderChoice(choice, isFocused))
		case app.InputBool:
			checkbox := m.checkboxes[field.index]
			b.WriteString(m.renderCheckbox(checkbox, isFocused))
		case app.InputFile:
			flagName := m.def.Flags[field.flagIndex].Name
			b.WriteString(m.renderFilePicker(&m.filePickers[field.index], flagName, isFocused))
		case app.InputDate:
			flagName := m.def.Flags[field.flagIndex].Name
			b.WriteString(m.renderDateTimePicker(m.datePickers[field.index], flagName, isFocused))
		default: // Text inputs
			b.WriteString(m.inputs[field.index].View())
		}
		b.WriteString("\n")
		if isFocused && (field.fieldType == app.InputLongString ||
			field.fieldType == app.InputList ||
			field.fieldType == app.InputFile ||
			field.fieldType == app.InputDate) {
			b.WriteString("\n")
		}
	}
}

func (m *formModel) renderValidationErrors(b *strings.Builder) {
	if m.validationErr != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.validationErr))
		b.WriteString("\n")
	}
}

func (m *formModel) renderSubmitButton(b *strings.Builder) {
	button := blurredButton
	currentVariant := m.variants[m.variantIndex]
	if m.focused == len(currentVariant) {
		button = focusedButton
	}
	b.WriteString("\n" + button + "\n")
}

func (m *formModel) renderHelp() string {
	useFullHelp := m.width == 0 || m.width < 120

	if useFullHelp {
		if len(m.variants) > 1 {
			return m.help.FullHelpView(m.keys.FullHelpWithVariants())
		}
		return m.help.FullHelpView(m.keys.FullHelp())
	}

	if len(m.variants) > 1 {
		return m.help.ShortHelpView(m.keys.ShortHelpWithVariants())
	}
	return m.help.ShortHelpView(m.keys.ShortHelp())
}
