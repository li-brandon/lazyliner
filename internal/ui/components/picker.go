package components

import (
	"github.com/brandonli/lazyliner/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PickerItem represents an item in the picker
type PickerItem struct {
	ID    string
	Label string
	Icon  string
	Desc  string
}

// PickerModel is a modal picker for selecting items
type PickerModel struct {
	title    string
	items    []PickerItem
	cursor   int
	selected string
	width    int
	height   int
}

// NewPickerModel creates a new picker model
func NewPickerModel(title string, items []PickerItem, width, height int) *PickerModel {
	return &PickerModel{
		title:  title,
		items:  items,
		cursor: 0,
		width:  width,
		height: height,
	}
}

// Update handles messages
func (m *PickerModel) Update(msg tea.Msg) (*PickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.items) > 0 {
				m.selected = m.items[m.cursor].ID
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.items) - 1
		}
	}
	return m, nil
}

// Selected returns the selected item ID
func (m *PickerModel) Selected() string {
	return m.selected
}

// SelectedItem returns the selected item
func (m *PickerModel) SelectedItem() *PickerItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return &m.items[m.cursor]
	}
	return nil
}

// View renders the picker
func (m *PickerModel) View() string {
	// Modal container
	modalWidth := 40
	modalHeight := len(m.items) + 4
	if modalHeight > m.height-4 {
		modalHeight = m.height - 4
	}

	// Title
	title := theme.ModalTitleStyle.Render(m.title)

	// Items
	var items string
	for i, item := range m.items {
		cursor := "  "
		style := theme.ListItemStyle
		if i == m.cursor {
			cursor = "> "
			style = theme.ListItemSelectedStyle
		}

		icon := ""
		if item.Icon != "" {
			icon = item.Icon + " "
		}

		items += style.Render(cursor + icon + item.Label) + "\n"
	}

	// Help
	help := theme.HelpStyle.Render("↑/↓: navigate  enter: select  esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, title, items, help)

	modal := theme.ModalStyle.
		Width(modalWidth).
		Render(content)

	// Center the modal
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}
