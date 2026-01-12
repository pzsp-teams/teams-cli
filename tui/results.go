package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var resultTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("205")).
	MarginBottom(1)

// resultsModel displays command execution results
type resultsModel struct {
	viewport viewport.Model
	ready    bool
	content  string
	title    string
	loading  bool
	spinner  spinner.Model
}

// NewResultsModel creates a new results display model
func NewResultsModel(title, content string) *resultsModel {
	return &resultsModel{
		title:   title,
		content: content,
		loading: false,
	}
}

// NewLoadingResultsModel creates a new results display model in loading state
func NewLoadingResultsModel(title string) *resultsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return &resultsModel{
		title:   title,
		loading: true,
		spinner: s,
	}
}

// Init implements tea.Model
func (m *resultsModel) Init() tea.Cmd {
	if m.loading {
		return m.spinner.Tick
	}
	return nil
}

// Update implements tea.Model
func (m *resultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
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
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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
	titleText := m.title
	if m.loading {
		titleText = fmt.Sprintf("%s %s", m.spinner.View(), titleText)
	}
	title := resultTitleStyle.Render(titleText)
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
