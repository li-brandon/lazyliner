package components

import (
	"strings"

	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/charmbracelet/bubbles/textinput"
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

// PickerModel is a modal picker for selecting items with search/filter support
type PickerModel struct {
	title         string
	items         []PickerItem
	filteredItems []PickerItem
	cursor        int
	selected      string
	width         int
	height        int
	searchInput   textinput.Model
	searchEnabled bool
}

// NewPickerModel creates a new picker model with search enabled by default
func NewPickerModel(title string, items []PickerItem, width, height int) *PickerModel {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 34

	return &PickerModel{
		title:         title,
		items:         items,
		filteredItems: items,
		cursor:        0,
		width:         width,
		height:        height,
		searchInput:   ti,
		searchEnabled: true,
	}
}

// NewPickerModelWithoutSearch creates a picker without search functionality
func NewPickerModelWithoutSearch(title string, items []PickerItem, width, height int) *PickerModel {
	return &PickerModel{
		title:         title,
		items:         items,
		filteredItems: items,
		cursor:        0,
		width:         width,
		height:        height,
		searchEnabled: false,
	}
}

// filterItems filters items based on search query
func (m *PickerModel) filterItems() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.filteredItems = m.items
		return
	}

	var filtered []PickerItem
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Label), query) ||
			strings.Contains(strings.ToLower(item.Desc), query) {
			filtered = append(filtered, item)
		}
	}
	m.filteredItems = filtered

	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filteredItems) {
		m.cursor = 0
	}
}

// Update handles messages
func (m *PickerModel) Update(msg tea.Msg) (*PickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "ctrl+j":
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.filteredItems) > 0 && m.cursor < len(m.filteredItems) {
				m.selected = m.filteredItems[m.cursor].ID
			}
		case "home", "ctrl+g":
			m.cursor = 0
		case "end":
			if len(m.filteredItems) > 0 {
				m.cursor = len(m.filteredItems) - 1
			}
		default:
			// Forward to search input if search is enabled
			if m.searchEnabled {
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.filterItems()
				return m, cmd
			}
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
	if m.cursor >= 0 && m.cursor < len(m.filteredItems) {
		return &m.filteredItems[m.cursor]
	}
	return nil
}

// View renders the picker
func (m *PickerModel) View() string {
	// Modal container
	modalWidth := 44
	maxVisibleItems := m.height - 10
	if maxVisibleItems < 5 {
		maxVisibleItems = 5
	}

	// Title
	title := theme.ModalTitleStyle.Render(m.title)

	var searchBar string
	if m.searchEnabled {
		searchBar = theme.InputStyle.Width(modalWidth - 6).Render(m.searchInput.View())
	}

	// Calculate scroll window
	startIdx := 0
	endIdx := len(m.filteredItems)
	if len(m.filteredItems) > maxVisibleItems {
		// Keep cursor centered when possible
		halfWindow := maxVisibleItems / 2
		if m.cursor > halfWindow {
			startIdx = m.cursor - halfWindow
		}
		endIdx = startIdx + maxVisibleItems
		if endIdx > len(m.filteredItems) {
			endIdx = len(m.filteredItems)
			startIdx = endIdx - maxVisibleItems
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Items
	var items string
	if len(m.filteredItems) == 0 {
		items = theme.TextMutedStyle.Render("  No matches found\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			item := m.filteredItems[i]
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

			items += style.Render(cursor+icon+item.Label) + "\n"
		}
	}

	// Scroll indicators
	var scrollIndicator string
	if len(m.filteredItems) > maxVisibleItems {
		if startIdx > 0 && endIdx < len(m.filteredItems) {
			scrollIndicator = theme.TextMutedStyle.Render("  ▲ ▼ scroll for more")
		} else if startIdx > 0 {
			scrollIndicator = theme.TextMutedStyle.Render("  ▲ scroll up for more")
		} else if endIdx < len(m.filteredItems) {
			scrollIndicator = theme.TextMutedStyle.Render("  ▼ scroll down for more")
		}
	}

	// Help text
	var helpText string
	if m.searchEnabled {
		helpText = "↑/↓: navigate  enter: select  esc: cancel"
	} else {
		helpText = "↑/↓: navigate  enter: select  esc: cancel"
	}
	help := theme.HelpStyle.Render(helpText)

	// Build content
	var contentParts []string
	contentParts = append(contentParts, title)
	if m.searchEnabled {
		contentParts = append(contentParts, searchBar, "")
	}
	contentParts = append(contentParts, items)
	if scrollIndicator != "" {
		contentParts = append(contentParts, scrollIndicator)
	}
	contentParts = append(contentParts, help)

	content := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

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
