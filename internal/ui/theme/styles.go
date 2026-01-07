package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Common styles used throughout the application
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(Background)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Surface).
			Padding(0, 1).
			Bold(true)

	LogoStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	HeaderInfoStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Padding(0, 2).
			Bold(true).
			Underline(true)

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(Surface).
			Padding(0, 1)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true)

	StatusBarDescStyle = lipgloss.NewStyle().
				Foreground(TextMuted)

	SearchBarStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(SurfaceHover).
			Padding(0, 1)

	// List styles
	ListItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(TextBright).
				Background(SurfaceHover).
				Padding(0, 1)

	ListItemDimStyle = lipgloss.NewStyle().
				Foreground(TextMuted).
				Padding(0, 1)

	// Issue ID style
	IssueIDStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Text).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	// Panel/Card styles
	PanelStyle = lipgloss.NewStyle().
			Background(Surface).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(1, 2)

	PanelTitleStyle = lipgloss.NewStyle().
			Foreground(Text).
			Bold(true).
			Padding(0, 0, 1, 0)

	// Modal styles
	ModalStyle = lipgloss.NewStyle().
			Background(Surface).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(Text).
			Bold(true).
			Padding(0, 0, 1, 0)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(BackgroundLight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(Text).
				Background(BackgroundLight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)

	// Label styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(TextBright).
			Background(SurfaceLight).
			Padding(0, 1).
			MarginRight(1)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(SurfaceLight).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(TextBright).
				Background(Primary).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryBright)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Bold(true)

	// Error/Warning styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Primary)

	// Divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(Border)

	// Text styles (for rendering colored text)
	TextMutedStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	TextDimStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	TextStyle = lipgloss.NewStyle().
			Foreground(Text)
)

// StatusStyle returns the appropriate style for a status
func StatusStyle(status string) lipgloss.Style {
	var color lipgloss.Color
	switch status {
	case "backlog", "Backlog":
		color = StatusBacklog
	case "todo", "Todo", "To Do":
		color = StatusTodo
	case "in_progress", "In Progress", "inProgress", "started":
		color = StatusInProgress
	case "done", "Done", "completed":
		color = StatusDone
	case "canceled", "Canceled", "cancelled":
		color = StatusCanceled
	default:
		color = TextMuted
	}
	return lipgloss.NewStyle().Foreground(color)
}

// StatusIcon returns an icon for a status
func StatusIcon(status string) string {
	switch status {
	case "backlog", "Backlog":
		return "📋"
	case "todo", "Todo", "To Do":
		return "⚪"
	case "in_progress", "In Progress", "inProgress", "started":
		return "🟡"
	case "done", "Done", "completed":
		return "🟢"
	case "canceled", "Canceled", "cancelled":
		return "🚫"
	default:
		return "○"
	}
}

func Divider(width int) string {
	return DividerStyle.Render(strings.Repeat("─", width))
}
