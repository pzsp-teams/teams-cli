package tui

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	bubbledatetimepicker "github.com/lcc/bubble-datetime-picker"

	"github.com/pzsp-teams/teams-cli/app"
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

// checkboxField represents a checkbox for InputBool
type checkboxField struct {
	flagDef *app.FlagDef
	checked bool
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

func createCheckbox(flagDef *app.FlagDef) checkboxField {
	checkbox := checkboxField{
		flagDef: flagDef,
		checked: false,
	}

	if flagDef.DefaultVal != nil {
		if defBool, ok := flagDef.DefaultVal.(bool); ok {
			checkbox.checked = defBool
		}
	}

	return checkbox
}

func createTextInput(flagDef *app.FlagDef) textinput.Model {
	t := textinput.New()
	t.CharLimit = 256
	t.Width = 50
	t.Prompt = flagDef.Name + ": "
	t.Placeholder = flagDef.Usage

	if flagDef.Type == app.InputPassword {
		t.EchoMode = textinput.EchoPassword
	}

	return t
}

func createFilePicker() filepicker.Model {
	fp := filepicker.New()
	fp.AllowedTypes = nil
	fp.CurrentDirectory = "."
	fp.ShowHidden = true
	fp.ShowPermissions = false
	fp.SetHeight(10)
	return fp
}

func createDateTimePicker() *bubbledatetimepicker.DateAndHourModel {
	dp := bubbledatetimepicker.NewDateAndHourModel()
	return &dp
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

// formComponents holds all the initialized UI components for a form
type formComponents struct {
	inputs      []textinput.Model
	choices     []choiceField
	checkboxes  []checkboxField
	textAreas   []textarea.Model
	filePickers []filepicker.Model
	datePickers []*bubbledatetimepicker.DateAndHourModel
	fieldMap    map[int]fieldInfo
}

func buildFields(def *app.CommandDef) formComponents {
	var (
		inputs      []textinput.Model
		choices     []choiceField
		checkboxes  []checkboxField
		textAreas   []textarea.Model
		filePickers []filepicker.Model
		datePickers []*bubbledatetimepicker.DateAndHourModel
		fieldMap    = make(map[int]fieldInfo)
	)

	for i := range def.Flags {
		f := &def.Flags[i]

		switch f.Type {
		case app.InputChoice:
			choices = append(choices, createChoiceField(f))
			fieldMap[i] = fieldInfo{app.InputChoice, len(choices) - 1}

		case app.InputBool:
			checkboxes = append(checkboxes, createCheckbox(f))
			fieldMap[i] = fieldInfo{app.InputBool, len(checkboxes) - 1}

		case app.InputLongString, app.InputList:
			textAreas = append(textAreas, createTextArea(f))
			fieldMap[i] = fieldInfo{f.Type, len(textAreas) - 1}

		case app.InputFile:
			filePickers = append(filePickers, createFilePicker())
			fieldMap[i] = fieldInfo{app.InputFile, len(filePickers) - 1}

		case app.InputDate:
			datePickers = append(datePickers, createDateTimePicker())
			fieldMap[i] = fieldInfo{app.InputDate, len(datePickers) - 1}

		default:
			inputs = append(inputs, createTextInput(f))
			fieldMap[i] = fieldInfo{f.Type, len(inputs) - 1}
		}
	}

	return formComponents{
		inputs:      inputs,
		choices:     choices,
		checkboxes:  checkboxes,
		textAreas:   textAreas,
		filePickers: filePickers,
		datePickers: datePickers,
		fieldMap:    fieldMap,
	}
}
