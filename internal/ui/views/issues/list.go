package issues

import (
	"fmt"
	"strings"

	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/brandonli/lazyliner/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// ListModel is the issue list view
type ListModel struct {
	issues   []linear.Issue
	cursor   int
	offset   int
	width    int
	height   int
	pageSize int
}

// NewListModel creates a new list model
func NewListModel(issues []linear.Issue, width, height int) ListModel {
	pageSize := height - 2
	if pageSize < 1 {
		pageSize = 10
	}
	return ListModel{
		issues:   issues,
		cursor:   0,
		offset:   0,
		width:    width,
		height:   height,
		pageSize: pageSize,
	}
}

// SetSize updates the list dimensions
func (m ListModel) SetSize(width, height int) ListModel {
	m.width = width
	m.height = height
	m.pageSize = height - 2
	if m.pageSize < 1 {
		m.pageSize = 10
	}
	return m
}

// Update handles messages
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.pageSize {
					m.offset = m.cursor - m.pageSize + 1
				}
			}
		case "home", "g":
			m.cursor = 0
			m.offset = 0
		case "end", "G":
			m.cursor = len(m.issues) - 1
			if m.cursor >= m.pageSize {
				m.offset = m.cursor - m.pageSize + 1
			}
		case "pgup", "ctrl+u":
			m.cursor -= m.pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.offset = m.cursor
		case "pgdown", "ctrl+d":
			m.cursor += m.pageSize
			if m.cursor >= len(m.issues) {
				m.cursor = len(m.issues) - 1
			}
			if m.cursor >= m.offset+m.pageSize {
				m.offset = m.cursor - m.pageSize + 1
			}
		}
	}
	return m, nil
}

// SelectedIssue returns the currently selected issue
func (m ListModel) SelectedIssue() *linear.Issue {
	if m.cursor >= 0 && m.cursor < len(m.issues) {
		return &m.issues[m.cursor]
	}
	return nil
}

// View renders the list
func (m ListModel) View() string {
	if len(m.issues) == 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			theme.TextMutedStyle.Render("No issues found"),
		)
	}

	var rows []string

	// Calculate visible range
	end := m.offset + m.pageSize
	if end > len(m.issues) {
		end = len(m.issues)
	}

	// Column widths
	idWidth := 10
	titleWidth := m.width - idWidth - 30 // Leave room for priority and status
	if titleWidth < 20 {
		titleWidth = 20
	}
	priorityWidth := 10
	statusWidth := 15

	for i := m.offset; i < end; i++ {
		issue := m.issues[i]
		isSelected := i == m.cursor

		row := m.renderRow(issue, isSelected, idWidth, titleWidth, priorityWidth, statusWidth)
		rows = append(rows, row)
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.issues) > m.pageSize {
		scrollInfo = theme.TextDimStyle.Render(fmt.Sprintf(" %d/%d ", m.cursor+1, len(m.issues)))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Add scroll info at bottom right
	if scrollInfo != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content,
			lipgloss.PlaceHorizontal(m.width, lipgloss.Right, scrollInfo))
	}

	return content
}

// renderRow renders a single issue row
func (m ListModel) renderRow(issue linear.Issue, isSelected bool, idWidth, titleWidth, priorityWidth, statusWidth int) string {
	baseStyle := theme.ListItemStyle
	if isSelected {
		baseStyle = theme.ListItemSelectedStyle
	}
	cursor := "○ "
	if isSelected {
		cursor = "● "
	} else {
		cursor = "○ "
	}

	// Issue ID
	id := theme.IssueIDStyle.Render(util.Truncate(issue.Identifier, idWidth))

	// Title
	title := util.Truncate(issue.Title, titleWidth)
	if isSelected {
		title = lipgloss.NewStyle().Foreground(theme.TextBright).Render(title)
	} else {
		title = lipgloss.NewStyle().Foreground(theme.Text).Render(title)
	}

	// Priority
	priorityIcon := theme.PriorityIcon(issue.Priority)
	priorityLabel := theme.PriorityLabel(issue.Priority)
	priority := lipgloss.NewStyle().
		Foreground(theme.PriorityColor(issue.Priority)).
		Width(priorityWidth).
		Render(priorityIcon + " " + priorityLabel)

	// Status
	var statusName string
	var statusType string
	if issue.State != nil {
		statusName = issue.State.Name
		statusType = issue.State.Type
	} else {
		statusName = "Unknown"
		statusType = ""
	}
	statusIcon := theme.StatusIcon(statusType)
	status := theme.StatusStyle(statusType).
		Width(statusWidth).
		Render(statusIcon + " " + util.Truncate(statusName, statusWidth-3))

	// Build row
	row := fmt.Sprintf("%s%s  %s  %s  %s",
		cursor,
		padRight(id, idWidth),
		padRight(title, titleWidth),
		priority,
		status,
	)

	return baseStyle.Width(m.width).Render(row)
}

func padRight(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}
