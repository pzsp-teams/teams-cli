package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
func (i item) FilterValue() string { return "" }

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

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + s[0])
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

// Init implements tea.Model interface
func (m *model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model interface
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model interface
func (m *model) View() string {
	if m.choice != "" {
		return fmt.Sprintf("You selected: %s\n", m.choice)
	}
	if m.quitting {
		return "Exiting...\n"
	}
	return "\n" + m.list.View()
}

func initialModel() *model {
	items := []list.Item{
		item("Teams - Manage teams"),
		item("Channels - Manage channels"),
		item("Chats - Manage chats"),
	}

	const defaultWidth = 80

	l := list.New(items, itemDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Teams CLI - Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &model{list: l}
}
