package setup

import (
	"github.com/brandonli/lazyliner/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the setup/welcome view model
type Model struct {
	width  int
	height int
}

// New creates a new setup model
func New(width, height int) Model {
	return Model{width: width, height: height}
}

// SetSize updates the view dimensions
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the setup view
func (m Model) View() string {
	// Welcome header
	logo := theme.LogoStyle.Render("Welcome to Lazyliner")

	subtitle := theme.TextMutedStyle.Render("A beautiful, keyboard-driven terminal TUI for Linear")

	// Setup instructions
	sections := []struct {
		title   string
		content string
	}{
		{
			title: "Step 1: Get your Linear API Key",
			content: `1. Go to Linear Settings > API
   https://linear.app/settings/api

2. Click "Create key" to generate a new API key

3. Copy the key (starts with lin_api_)`,
		},
		{
			title: "Step 2: Configure Lazyliner",
			content: `Option A: Environment Variable (recommended)

   export LAZYLINER_API_KEY=lin_api_xxxxx

   Add this to your shell profile (.bashrc, .zshrc, etc.)


Option B: Config File

   Create ~/.config/lazyliner/config.yaml:

   linear:
     api_key: lin_api_xxxxx`,
		},
		{
			title: "Step 3: Launch Lazyliner",
			content: `Once configured, simply run:

   lazyliner

   Press ? for help and keyboard shortcuts`,
		},
	}

	// Render sections
	var renderedSections []string
	for _, section := range sections {
		titleStyle := lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			MarginBottom(1)

		contentStyle := lipgloss.NewStyle().
			Foreground(theme.Text).
			MarginLeft(2)

		rendered := lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render(section.title),
			contentStyle.Render(section.content),
		)
		renderedSections = append(renderedSections, rendered)
	}

	// Footer hint
	footer := theme.TextDimStyle.Render("Press q to quit")

	// Combine all content
	contentWidth := min(m.width-8, 70)

	contentBox := lipgloss.NewStyle().
		Width(contentWidth).
		Padding(1, 2)

	inner := lipgloss.JoinVertical(
		lipgloss.Center,
		logo,
		subtitle,
		"",
		contentBox.Render(lipgloss.JoinVertical(lipgloss.Left, renderedSections...)),
		"",
		footer,
	)

	// Center the entire content
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		inner,
	)
}
