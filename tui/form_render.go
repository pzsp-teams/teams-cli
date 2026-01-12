package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	bubbledatetimepicker "github.com/lcc/bubble-datetime-picker"

	"github.com/pzsp-teams/teams-cli/app"
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

func (m *formModel) renderCheckbox(checkbox checkboxField, focused bool) string {
	var b strings.Builder

	checkMark := "[ ]"
	if checkbox.checked {
		checkMark = "[x]"
	}

	if focused {
		b.WriteString(focusedStyle.Render(checkMark + " " + checkbox.flagDef.Name))
	} else {
		b.WriteString(noStyle.Render(checkMark + " " + checkbox.flagDef.Name))
	}

	if checkbox.flagDef.Usage != "" {
		b.WriteString(blurredStyle.Render(" - " + checkbox.flagDef.Usage))
	}

	return b.String()
}

func (m *formModel) renderTextArea(ta *textarea.Model, name string, focused bool) string {
	var b strings.Builder

	labelStyle := noStyle
	if focused {
		labelStyle = focusedStyle
	}

	b.WriteString(labelStyle.Render(name + ":"))
	b.WriteString("\n")

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

func (m *formModel) renderFilePicker(fp *filepicker.Model, name string, focused bool) string {
	var b strings.Builder

	labelStyle := noStyle
	if focused {
		labelStyle = focusedStyle
	}

	b.WriteString(labelStyle.Render(name + ":"))
	b.WriteString("\n")

	if focused {
		// Add instructions
		helpText := blurredStyle.Render("(j/k: up/down • g/G: first/last • h/l: back/enter • Enter: select)")
		b.WriteString(helpText + "\n")
		b.WriteString(fp.View())
	} else {
		// Collapsed view - show selected file
		selectedFile := fp.Path
		if selectedFile == "" {
			selectedFile = "(no file selected)"
		}
		b.WriteString(noStyle.Render("> " + selectedFile))
	}
	return b.String()
}

func (m *formModel) renderDateTimePicker(dp *bubbledatetimepicker.DateAndHourModel, name string, focused bool) string {
	var b strings.Builder

	labelStyle := noStyle
	if focused {
		labelStyle = focusedStyle
	}

	b.WriteString(labelStyle.Render(name + ":"))
	b.WriteString("\n")

	if focused {
		// Add instructions
		helpText := blurredStyle.Render("(↑↓←→: navigate • Tab: next component • Enter: next field • Del: back)")
		b.WriteString(helpText + "\n")
		b.WriteString(dp.View())
	} else {
		// Collapsed view - show selected date/time
		dateTime := dp.Time()
		b.WriteString(noStyle.Render("> " + dateTime.Format("2006-01-02 15:04")))
	}
	return b.String()
}

func isTextInput(t app.InputType) bool {
	return t != app.InputChoice &&
		t != app.InputBool &&
		t != app.InputLongString &&
		t != app.InputList &&
		t != app.InputFile &&
		t != app.InputDate &&
		t != app.InputNone
}
