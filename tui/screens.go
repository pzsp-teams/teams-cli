package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/pzsp-teams/teams-cli/app"
)

// CreateMenuList creates a list model for a parent command definition
func CreateMenuList(registry []app.CommandDef, def *app.CommandDef) list.Model {
	var items []list.Item
	var title string

	if def == nil {
		title = "Teams CLI - Main Menu"
		for i := range registry {
			items = append(items, item(registry[i].Short))
		}
	} else {
		title = def.Short
		for i := range def.SubCommands {
			d := &def.SubCommands[i]
			label := d.Short
			if label == "" {
				label = d.Use
			}
			items = append(items, item(label))
		}
		items = append(items, item("Back"))
	}

	const defaultWidth = 80

	l := list.New(items, itemDelegate{}, defaultWidth, defaultHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}
