package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"

	"github.com/pzsp-teams/cli/app"
)

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

func (m *formModel) renderTextArea(ta *textarea.Model, name string, focused bool) string {
	var b strings.Builder

	labelStyle := noStyle
	if focused {
		labelStyle = focusedStyle
	}

	b.WriteString(labelStyle.Render(name + ":\n"))

	if focused {
		b.WriteString(ta.View())
	} else {
		// Collapsed view
		val := ta.Value()
		lines := strings.Split(val, "\n")
		display := ""
		if len(lines) > 0 {
			display = lines[0]
			if len(lines) > 1 || len(lines[0]) > 47 {
				if len(display) > 47 {
					display = display[:47]
				}
				display += "..."
			}
		}
		if display == "" {
			display = ta.Placeholder
		}

		// Simulate a text input look
		b.WriteString(noStyle.Render("> " + display))
	}
	return b.String()
}

func isTextInput(t app.InputType) bool {
	return t != app.InputChoice && t != app.InputLongString && t != app.InputList && t != app.InputNone
}
