package help

import (
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width  int
	height int
}

func New(width, height int) Model {
	return Model{width: width, height: height}
}

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

func (m Model) View() string {
	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			title: "Navigation",
			keys: [][]string{
				{"j / ↓", "Move down"},
				{"k / ↑", "Move up"},
				{"g / Home", "Go to top"},
				{"G / End", "Go to bottom"},
				{"Ctrl+d", "Page down"},
				{"Ctrl+u", "Page up"},
			},
		},
		{
			title: "Tabs",
			keys: [][]string{
				{"Tab", "Next tab"},
				{"Shift+Tab", "Previous tab"},
				{"1", "My Issues"},
				{"2", "All Issues"},
				{"3", "Active"},
				{"4", "Backlog"},
			},
		},
		{
			title: "Actions",
			keys: [][]string{
				{"Enter", "View issue"},
				{"/", "Search issues"},
				{"c", "Create issue"},
				{"r", "Refresh"},
				{"Esc", "Back / Cancel"},
				{"q", "Quit"},
			},
		},
		{
			title: "Issue Actions",
			keys: [][]string{
				{"s", "Change status"},
				{"a", "Change assignee"},
				{"p", "Change priority"},
				{"y", "Copy branch name"},
				{"o", "Open in browser"},
			},
		},
	}

	var cols []string
	for _, section := range sections {
		col := renderSection(section.title, section.keys)
		cols = append(cols, col)
	}

	left := lipgloss.JoinVertical(lipgloss.Left, cols[0], cols[1])
	right := lipgloss.JoinVertical(lipgloss.Left, cols[2], cols[3])
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", right)

	title := theme.ModalTitleStyle.Render("Keyboard Shortcuts")
	hint := theme.TextDimStyle.Render("Press ? or Esc to close")

	inner := lipgloss.JoinVertical(lipgloss.Center, title, "", content, "", hint)

	modal := theme.ModalStyle.
		Width(min(m.width-4, 70)).
		Render(inner)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func renderSection(title string, keys [][]string) string {
	header := theme.TitleStyle.Render(title)
	var rows []string
	for _, kv := range keys {
		key := theme.StatusBarKeyStyle.Width(12).Render(kv[0])
		desc := theme.TextMutedStyle.Render(kv[1])
		rows = append(rows, key+desc)
	}
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.JoinVertical(lipgloss.Left, header, content, "")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
