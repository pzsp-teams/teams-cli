package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/pzsp-teams/cli/app"
)

// formField tracks a field in the form and its type
type formField struct {
	fieldType app.InputType
	index     int // Index in inputs, choices, or textareas array
	flagIndex int // Index in the CommandDef.Flags array
}

// choiceField represents a radio button group for InputChoice
type choiceField struct {
	flagDef  *app.FlagDef
	selected int
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

func createTextInput(flagDef *app.FlagDef) textinput.Model {
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

	return t
}

func createTextArea(flagDef *app.FlagDef) textarea.Model {
	t := textarea.New()
	t.Placeholder = flagDef.Usage
	if flagDef.Type == app.InputList {
		t.Placeholder = "Enter items (one per line)"
	}
	t.Prompt = ""
	t.CharLimit = 0
	t.ShowLineNumbers = false
	t.SetWidth(50)
	t.SetHeight(5)
	return t
}

func buildFields(def *app.CommandDef) (
	inputs []textinput.Model,
	choices []choiceField,
	textAreas []textarea.Model,
	fieldMap map[int]fieldInfo,
) {
	fieldMap = make(map[int]fieldInfo)

	for i := range def.Flags {
		f := &def.Flags[i]

		switch f.Type {
		case app.InputChoice:
			choices = append(choices, createChoiceField(f))
			fieldMap[i] = fieldInfo{app.InputChoice, len(choices) - 1}

		case app.InputLongString, app.InputList:
			textAreas = append(textAreas, createTextArea(f))
			fieldMap[i] = fieldInfo{f.Type, len(textAreas) - 1}

		default:
			inputs = append(inputs, createTextInput(f))
			fieldMap[i] = fieldInfo{f.Type, len(inputs) - 1}
		}
	}

	return
}
