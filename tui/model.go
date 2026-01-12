package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/pzsp-teams/cli/app"
)

const defaultHeight = 24

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item string

// FilterValue implements list.Item interface
func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

// Height implements list.ItemDelegate interface
func (d itemDelegate) Height() int { return 1 }

// Spacing implements list.ItemDelegate interface
func (d itemDelegate) Spacing() int { return 0 }

// Update implements list.ItemDelegate interface
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render implements list.ItemDelegate interface
//
//nolint:gocritic // m list.Model must be passed by value per interface requirement
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + s[0])
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type mode int

const (
	modeMenu mode = iota
	modeForm
	modeResults
)

type model struct {
	registry    []app.CommandDef
	mode        mode
	list        list.Model
	form        *formModel
	results     *resultsModel
	messageList *messageListModel
	navigation  *navigationState
	executor    *CommandExecutor
	width       int
	height      int
	quitting    bool
}

// Init implements tea.Model interface
func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model interface
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.KeyMsg:
		if model, cmd, handled := m.handleGlobalKeys(msg); handled {
			return model, cmd
		}
	}

	switch m.mode {
	case modeForm:
		return m.updateForm(msg)
	case modeResults:
		return m.updateResults(msg)
	default:
		return m.updateList(msg)
	}
}

func (m *model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.list.SetSize(msg.Width, msg.Height)
	if m.mode == modeResults {
		var cmd tea.Cmd
		_, cmd = m.results.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit, true

	case "q":
		if m.mode == modeMenu {
			m.quitting = true
			return m, tea.Quit, true
		}
		if m.mode == modeResults {
			m.exitResultsMode()
			return m, nil, true
		}

	case "esc":
		if m.mode == modeForm {
			m.mode = modeMenu
			return m, nil, true
		}
		if m.mode == modeResults {
			m.exitResultsMode()
			return m, nil, true
		}
		if m.navigation.goBack() {
			m.updateListForCurrentScreen()
			return m, nil, true
		}
		m.quitting = true
		return m, tea.Quit, true
	}
	return m, nil, false
}

func (m *model) exitResultsMode() {
	if m.form != nil {
		m.mode = modeForm
	} else {
		m.mode = modeMenu
	}
}

func (m *model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		return m.handleSelection()
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentVariantIdx := m.form.variantIndex
	if currentVariantIdx >= len(m.form.variants) {
		currentVariantIdx = 0
	}
	variantSize := len(m.form.variants[currentVariantIdx])

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" && m.form.focused == variantSize {
		return m.executeForm()
	}
	var cmd tea.Cmd
	_, cmd = m.form.Update(msg)
	return m, cmd
}

func (m *model) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	_, cmd = m.results.Update(msg)
	return m, cmd
}

func (m *model) executeForm() (tea.Model, tea.Cmd) {
	flags := make(map[string]any)

	currentVariantIdx := m.form.variantIndex
	if currentVariantIdx >= len(m.form.variants) {
		currentVariantIdx = 0
	}
	currentVariant := m.form.variants[currentVariantIdx]

	for _, field := range currentVariant {
		flagDef := &m.form.def.Flags[field.flagIndex]

		switch field.fieldType {
		case app.InputLongString:
			val := m.form.textAreas[field.index].Value()
			if val != "" {
				flags[flagDef.Name] = val
			}
		case app.InputList:
			val := m.form.textAreas[field.index].Value()
			if val != "" {
				lines := strings.Split(val, "\n")
				var filtered []string
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						filtered = append(filtered, trimmed)
					}
				}
				if len(filtered) > 0 {
					flags[flagDef.Name] = filtered
				}
			}
		case app.InputChoice:
			choice := m.form.choices[field.index]
			selectedValue := choice.flagDef.Options[choice.selected]
			flags[flagDef.Name] = selectedValue
		default: // Text inputs
			val := m.form.inputs[field.index].Value()
			if val != "" {
				flags[flagDef.Name] = val
			}
		}
	}

	output, _, err := m.executor.ExecuteCommand(m.form.def, flags)
	title := "Results: " + m.form.def.Use
	if err != nil {
		output = fmt.Sprintf("Error: %v\n\n%s", err, output)
	}

	m.results = NewResultsModel(title, output)
	m.mode = modeResults

	return m, func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}
}

func (m *model) handleSelection() (tea.Model, tea.Cmd) {
	selectedItem, ok := m.list.SelectedItem().(item)
	if !ok {
		return m, nil
	}

	selection := string(selectedItem)
	if selection == "Back" {
		m.navigation.goBack()
		m.updateListForCurrentScreen()
		return m, nil
	}

	selectedDef := m.findCommandDef(selection)
	if selectedDef == nil {
		return m, nil
	}

	if len(selectedDef.SubCommands) > 0 {
		m.navigation.navigateTo(selectedDef)
		m.updateListForCurrentScreen()
		return m, nil
	}

	if len(selectedDef.Flags) > 0 {
		m.form = NewFormModel(selectedDef)
		m.mode = modeForm
		return m, nil
	}

	return m.executeCommand(selectedDef)
}

func (m *model) findCommandDef(selection string) *app.CommandDef {
	currentDef := m.navigation.getCurrentCommand()
	var defs []app.CommandDef

	if currentDef == nil {
		defs = m.registry
	} else {
		defs = currentDef.SubCommands
	}

	for i := range defs {
		label := defs[i].Short
		if label == "" {
			label = defs[i].Use
		}
		if label == selection {
			return &defs[i]
		}
	}
	return nil
}

func (m *model) executeCommand(def *app.CommandDef) (tea.Model, tea.Cmd) {
	m.form = nil
	output, _, err := m.executor.ExecuteCommand(def, nil)
	title := "Results: " + def.Use
	if err != nil {
		output = fmt.Sprintf("Error: %v\n\n%s", err, output)
	}

	m.results = NewResultsModel(title, output)
	m.mode = modeResults
	return m, func() tea.Msg {
		return tea.WindowSizeMsg{Width: m.width, Height: m.height}
	}
}

func (m *model) updateListForCurrentScreen() {
	m.list = CreateMenuList(m.registry, m.navigation.getCurrentCommand())
	m.list.SetSize(m.width, m.height)
}

// View implements tea.Model interface
func (m *model) View() string {
	if m.quitting {
		return "Exiting...\n"
	}

	switch m.mode {
	case modeForm:
		return "\n" + m.form.View()
	case modeResults:
		return m.results.View()
	default:
		return "\n" + m.list.View()
	}
}

func initialModel(ctx context.Context, registry []app.CommandDef, startPath string) *model {
	executor := NewCommandExecutor(ctx)
	nav := newNavigationState()

	if startPath != "" {
		for i := range registry {
			if registry[i].Use == startPath {
				nav.navigateTo(&registry[i])
				break
			}
		}
	}

	m := &model{
		registry:   registry,
		mode:       modeMenu,
		list:       CreateMenuList(registry, nav.getCurrentCommand()),
		navigation: nav,
		executor:   executor,
	}
	return m
}
