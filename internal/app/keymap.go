package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Top      key.Binding
	Bottom   key.Binding
	PageUp   key.Binding
	PageDown key.Binding

	// Selection
	Enter  key.Binding
	Select key.Binding

	// Tab navigation
	NextTab key.Binding
	PrevTab key.Binding
	Tab1    key.Binding
	Tab2    key.Binding
	Tab3    key.Binding
	Tab4    key.Binding

	// Actions
	Create   key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Filter   key.Binding
	Help     key.Binding
	Quit     key.Binding
	Back     key.Binding
	Cancel   key.Binding

	// Issue actions
	Status   key.Binding
	Assignee key.Binding
	Priority key.Binding
	Labels   key.Binding
	Project  key.Binding

	// Utility
	CopyBranch    key.Binding
	OpenInBrowser key.Binding
	Comment       key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation (vim-style + arrows)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u", "pgup"),
			key.WithHelp("ctrl+u", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d", "pgdown"),
			key.WithHelp("ctrl+d", "page down"),
		),

		// Selection
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),

		// Tab navigation
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Tab1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "my issues"),
		),
		Tab2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "all issues"),
		),
		Tab3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "active"),
		),
		Tab4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "backlog"),
		),

		// Actions
		Create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "delete"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "cancel"),
		),

		// Issue actions
		Status: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "status"),
		),
		Assignee: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "assignee"),
		),
		Priority: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "priority"),
		),
		Labels: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "labels"),
		),
		Project: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "project"),
		),

		// Utility
		CopyBranch: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy branch"),
		),
		OpenInBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Comment: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "comment"),
		),
	}
}

// ShortHelp returns a short help string for the status bar
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Enter, k.Create, k.Status, k.Help, k.Quit,
	}
}

// FullHelp returns the full help for the help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.Top, k.Bottom, k.PageUp, k.PageDown},
		// Tabs
		{k.NextTab, k.PrevTab, k.Tab1, k.Tab2, k.Tab3, k.Tab4},
		// Actions
		{k.Enter, k.Create, k.Edit, k.Delete, k.Refresh, k.Search},
		// Issue actions
		{k.Status, k.Assignee, k.Priority, k.Labels, k.CopyBranch, k.OpenInBrowser},
		// General
		{k.Help, k.Back, k.Quit},
	}
}
