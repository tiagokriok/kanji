package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit            key.Binding
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	OpenDetails     key.Binding
	ToggleDetails   key.Binding
	NewTask         key.Binding
	EditTitle       key.Binding
	EditDescription key.Binding
	AddComment      key.Binding
	Search          key.Binding
	ClearSearch     key.Binding
	ToggleView      key.Binding
	MoveTask        key.Binding
	CycleStatus     key.Binding
	ToggleDueSoon   key.Binding
	Confirm         key.Binding
	Cancel          key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Quit:            key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Up:              key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:            key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:            key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:           key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		OpenDetails:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open/move")),
		ToggleDetails:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "toggle details")),
		NewTask:         key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new task")),
		EditTitle:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit title")),
		EditDescription: key.NewBinding(key.WithKeys("E"), key.WithHelp("E", "edit description")),
		AddComment:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "add comment")),
		Search:          key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		ClearSearch:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "clear search")),
		ToggleView:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch view")),
		MoveTask:        key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move task")),
		CycleStatus:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle column filter")),
		ToggleDueSoon:   key.NewBinding(key.WithKeys("z"), key.WithHelp("z", "due soon")),
		Confirm:         key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel:          key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}
