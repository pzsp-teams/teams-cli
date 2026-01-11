package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pzsp-teams/cli/internal/formatters"
)

type messageItem struct {
	msg *formatters.MessageView
}

func (i messageItem) Title() string       { return i.msg.Author }
func (i messageItem) Description() string { return i.msg.Content }
func (i messageItem) FilterValue() string { return i.msg.Author + " " + i.msg.Content }

type messageListModel struct {
	list     list.Model
	messages []*formatters.MessageView
	selected *formatters.MessageView
}

func NewMessageListModel(messages []*formatters.MessageView) *messageListModel {
	items := make([]list.Item, len(messages))
	for i, m := range messages {
		items[i] = messageItem{msg: m}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Messages"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return &messageListModel{
		list:     l,
		messages: messages,
	}
}

func (m *messageListModel) Init() tea.Cmd {
	return nil
}

func (m *messageListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(messageItem); ok {
				m.selected = i.msg
				// TODO: Return a specialized message to parent model to handle selection
				// For now, we just update selected state
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *messageListModel) View() string {
	return m.list.View()
}
