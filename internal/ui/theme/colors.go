package theme

import "github.com/charmbracelet/lipgloss"

// Linear-inspired color palette
var (
	// Primary colors
	Primary       = lipgloss.Color("#5E6AD2") // Linear purple
	PrimaryDim    = lipgloss.Color("#4E5ABF")
	PrimaryBright = lipgloss.Color("#7C85E3")

	// Secondary colors
	Secondary    = lipgloss.Color("#6B6F76") // Muted gray
	SecondaryDim = lipgloss.Color("#505459")

	// Status colors
	Success = lipgloss.Color("#4CA154") // Green
	Warning = lipgloss.Color("#D9730D") // Orange
	Danger  = lipgloss.Color("#EB5757") // Red
	Info    = lipgloss.Color("#2F81F7") // Blue

	// Background colors
	Background      = lipgloss.Color("#1C1C1C") // Dark background
	BackgroundLight = lipgloss.Color("#242424")
	Surface         = lipgloss.Color("#2C2C2C") // Card/panel background
	SurfaceLight    = lipgloss.Color("#363636")
	SurfaceHover    = lipgloss.Color("#404040")

	// Text colors
	Text       = lipgloss.Color("#FFFFFF")
	TextMuted  = lipgloss.Color("#6B6F76")
	TextDim    = lipgloss.Color("#505459")
	TextBright = lipgloss.Color("#FAFAFA")

	// Border colors
	Border      = lipgloss.Color("#363636")
	BorderFocus = lipgloss.Color("#5E6AD2")

	// Priority colors (matching Linear)
	PriorityUrgent = lipgloss.Color("#F2555A") // Red
	PriorityHigh   = lipgloss.Color("#F2994A") // Orange
	PriorityMedium = lipgloss.Color("#5E6AD2") // Purple (primary)
	PriorityLow    = lipgloss.Color("#6B6F76") // Gray
	PriorityNone   = lipgloss.Color("#505459") // Dim gray

	// Status colors (matching Linear workflow)
	StatusBacklog    = lipgloss.Color("#6B6F76") // Gray
	StatusTodo       = lipgloss.Color("#E2E2E2") // Light gray
	StatusInProgress = lipgloss.Color("#F2C94C") // Yellow
	StatusDone       = lipgloss.Color("#4CA154") // Green
	StatusCanceled   = lipgloss.Color("#6B6F76") // Gray
)

// PriorityColor returns the color for a given priority level (0-4)
func PriorityColor(priority int) lipgloss.Color {
	switch priority {
	case 1:
		return PriorityUrgent
	case 2:
		return PriorityHigh
	case 3:
		return PriorityMedium
	case 4:
		return PriorityLow
	default:
		return PriorityNone
	}
}

// PriorityLabel returns a human-readable label for priority
func PriorityLabel(priority int) string {
	switch priority {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "None"
	}
}

// PriorityIcon returns an icon for priority
func PriorityIcon(priority int) string {
	switch priority {
	case 1:
		return "🔴"
	case 2:
		return "🟠"
	case 3:
		return "🔵"
	case 4:
		return "⚪"
	default:
		return "◽"
	}
}
