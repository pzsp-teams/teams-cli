package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	resultTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	resultStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

// resultsModel displays command execution results
type resultsModel struct {
	viewport viewport.Model
	ready    bool
	content  string
	title    string
}

// NewResultsModel creates a new results display model
func NewResultsModel(title, content string) *resultsModel {
	return &resultsModel{
		title:   title,
		content: content,
	}
}

// Init implements tea.Model
func (m *resultsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *resultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m *resultsModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return strings.Join([]string{
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	}, "\n")
}

func (m *resultsModel) headerView() string {
	title := resultTitleStyle.Render(m.title)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *resultsModel) footerView() string {
	scrollPercent := fmt.Sprintf("%.0f%%", m.viewport.ScrollPercent()*100)
	info := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(scrollPercent)

	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info) +
		"\n  Press 'q' or 'esc' to return"
}
