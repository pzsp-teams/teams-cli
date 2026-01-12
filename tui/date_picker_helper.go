package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type datePickerDoneMsg struct{}

func wrapDatePickerCmd(cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		if cmd == nil {
			return nil
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			return datePickerDoneMsg{}
		}
		return msg
	}
}
