package issues

import (
	"fmt"
	"strings"
	"time"

	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailModel is the issue detail view
type DetailModel struct {
	issue      *linear.Issue
	width      int
	height     int
	scrollY    int
	maxScrollY int
}

// NewDetailModel creates a new detail model
func NewDetailModel(issue *linear.Issue, width, height int) DetailModel {
	return DetailModel{
		issue:   issue,
		width:   width,
		height:  height,
		scrollY: 0,
	}
}

// SetSize updates the detail view dimensions
func (m DetailModel) SetSize(width, height int) DetailModel {
	m.width = width
	m.height = height
	return m
}

// Update handles messages
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "down", "j":
			if m.scrollY < m.maxScrollY {
				m.scrollY++
			}
		case "home", "g":
			m.scrollY = 0
		case "end", "G":
			m.scrollY = m.maxScrollY
		}
	}
	return m, nil
}

// View renders the detail view
func (m DetailModel) View() string {
	if m.issue == nil {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			theme.TextMutedStyle.Render("No issue selected"),
		)
	}

	// Header with back button and ID
	header := m.renderHeader()

	// Title
	title := theme.TitleStyle.
		Width(m.width - 4).
		Render(m.issue.Title)

	// Metadata section
	metadata := m.renderMetadata()

	// Divider
	divider := theme.Divider(m.width - 4)

	// Description
	description := m.renderDescription()

	// Labels
	labels := m.renderLabels()

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		title,
		"",
		metadata,
		"",
		divider,
		"",
		description,
	)

	if labels != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", labels)
	}

	// Apply padding
	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(content)
}

// renderHeader renders the detail header
func (m DetailModel) renderHeader() string {
	back := theme.TextMutedStyle.Render("← ESC")
	id := theme.IssueIDStyle.Render(m.issue.Identifier)

	spacing := m.width - lipgloss.Width(back) - lipgloss.Width(id) - 8
	if spacing < 0 {
		spacing = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		back,
		strings.Repeat(" ", spacing),
		id,
	)
}

// renderMetadata renders the metadata section
func (m DetailModel) renderMetadata() string {
	var parts []string

	// Status
	if m.issue.State != nil {
		statusIcon := theme.StatusIcon(m.issue.State.Type)
		statusLabel := theme.StatusStyle(m.issue.State.Type).Render(m.issue.State.Name)
		parts = append(parts, fmt.Sprintf("Status: %s %s", statusIcon, statusLabel))
	}

	// Priority
	priorityIcon := theme.PriorityIcon(m.issue.Priority)
	priorityLabel := lipgloss.NewStyle().
		Foreground(theme.PriorityColor(m.issue.Priority)).
		Render(theme.PriorityLabel(m.issue.Priority))
	parts = append(parts, fmt.Sprintf("Priority: %s %s", priorityIcon, priorityLabel))

	// Assignee
	assignee := "Unassigned"
	if m.issue.Assignee != nil {
		assignee = m.issue.Assignee.Name
	}
	parts = append(parts, fmt.Sprintf("Assignee: %s", assignee))

	// Project
	if m.issue.Project != nil {
		parts = append(parts, fmt.Sprintf("Project: %s", m.issue.Project.Name))
	}

	// Team
	if m.issue.Team != nil {
		parts = append(parts, fmt.Sprintf("Team: %s", m.issue.Team.Name))
	}

	// Due date
	if m.issue.DueDate != nil && *m.issue.DueDate != "" {
		parts = append(parts, fmt.Sprintf("Due: %s", *m.issue.DueDate))
	}

	// Created/Updated
	parts = append(parts, fmt.Sprintf("Created: %s", formatRelativeTime(m.issue.CreatedAt)))
	parts = append(parts, fmt.Sprintf("Updated: %s", formatRelativeTime(m.issue.UpdatedAt)))

	// Render in two columns
	leftCol := []string{}
	rightCol := []string{}
	for i, p := range parts {
		if i%2 == 0 {
			leftCol = append(leftCol, p)
		} else {
			rightCol = append(rightCol, p)
		}
	}

	colWidth := (m.width - 8) / 2
	left := lipgloss.NewStyle().Width(colWidth).Render(strings.Join(leftCol, "\n"))
	right := lipgloss.NewStyle().Width(colWidth).Render(strings.Join(rightCol, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

// renderDescription renders the description
func (m DetailModel) renderDescription() string {
	if m.issue.Description == "" {
		return theme.TextMutedStyle.Render("No description")
	}

	// Simple markdown-ish rendering
	desc := m.issue.Description

	// Wrap to width
	maxWidth := m.width - 8
	if maxWidth < 40 {
		maxWidth = 40
	}

	wrapped := wordWrap(desc, maxWidth)
	return lipgloss.NewStyle().
		Foreground(theme.Text).
		Width(maxWidth).
		Render(wrapped)
}

// renderLabels renders the labels section
func (m DetailModel) renderLabels() string {
	if len(m.issue.Labels) == 0 {
		return ""
	}

	var labelStrs []string
	for _, label := range m.issue.Labels {
		// Use label color if available, otherwise default
		style := theme.LabelStyle
		if label.Color != "" {
			style = style.Background(lipgloss.Color(label.Color))
		}
		labelStrs = append(labelStrs, style.Render(label.Name))
	}

	return "Labels: " + strings.Join(labelStrs, " ")
}

// formatRelativeTime formats a time as relative (e.g., "2 hours ago")
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}
