package issues

import (
	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CreateModel is the issue creation form
type CreateModel struct {
	// Form fields
	titleInput textinput.Model
	descInput  textarea.Model

	// Options
	teams    []linear.Team
	projects []linear.Project
	states   []linear.WorkflowState
	users    []linear.User
	labels   []linear.Label

	// Selected values
	selectedTeam     int
	selectedProject  int
	selectedPriority int
	selectedAssignee int

	// UI state
	focusIndex int
	width      int
	height     int
}

// Field indices
const (
	fieldTitle = iota
	fieldDescription
	fieldTeam
	fieldProject
	fieldPriority
	fieldAssignee
	fieldCount
)

// NewCreateModel creates a new create model
func NewCreateModel(teams []linear.Team, projects []linear.Project, states []linear.WorkflowState, users []linear.User, labels []linear.Label, width, height int) CreateModel {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Issue title"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = width - 20

	// Description input
	ta := textarea.New()
	ta.Placeholder = "Description (markdown supported)"
	ta.CharLimit = 10000
	ta.SetWidth(width - 20)
	ta.SetHeight(6)

	return CreateModel{
		titleInput:       ti,
		descInput:        ta,
		teams:            teams,
		projects:         projects,
		states:           states,
		users:            users,
		labels:           labels,
		selectedTeam:     0,
		selectedProject:  -1, // No project by default
		selectedPriority: 0,  // No priority by default
		selectedAssignee: -1, // Unassigned by default
		focusIndex:       fieldTitle,
		width:            width,
		height:           height,
	}
}

// SetSize updates the form dimensions
func (m CreateModel) SetSize(width, height int) CreateModel {
	m.width = width
	m.height = height
	// Guard against uninitialized model (SetSize may be called before NewCreateModel)
	if m.titleInput.Placeholder != "" {
		m.titleInput.Width = width - 20
	}
	if m.descInput.Placeholder != "" {
		m.descInput.SetWidth(width - 20)
	}
	return m
}

// Update handles messages
func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % fieldCount
			m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + fieldCount) % fieldCount
			m.updateFocus()
		case "left":
			m.handleLeftRight(-1)
		case "right":
			m.handleLeftRight(1)
		default:
			// Forward to focused field
			switch m.focusIndex {
			case fieldTitle:
				var cmd tea.Cmd
				m.titleInput, cmd = m.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			case fieldDescription:
				var cmd tea.Cmd
				m.descInput, cmd = m.descInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// updateFocus updates which field is focused
func (m *CreateModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()

	switch m.focusIndex {
	case fieldTitle:
		m.titleInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	}
}

// handleLeftRight handles left/right navigation for select fields
func (m *CreateModel) handleLeftRight(dir int) {
	switch m.focusIndex {
	case fieldTeam:
		m.selectedTeam = clamp(m.selectedTeam+dir, 0, len(m.teams)-1)
	case fieldProject:
		m.selectedProject = clamp(m.selectedProject+dir, -1, len(m.projects)-1)
	case fieldPriority:
		m.selectedPriority = clamp(m.selectedPriority+dir, 0, 4)
	case fieldAssignee:
		m.selectedAssignee = clamp(m.selectedAssignee+dir, -1, len(m.users)-1)
	}
}

// GetInput returns the current form input as IssueCreateInput
func (m CreateModel) GetInput() linear.IssueCreateInput {
	input := linear.IssueCreateInput{
		Title:       m.titleInput.Value(),
		Description: m.descInput.Value(),
	}

	if m.selectedTeam >= 0 && m.selectedTeam < len(m.teams) {
		input.TeamID = m.teams[m.selectedTeam].ID
	}

	if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		input.ProjectID = m.projects[m.selectedProject].ID
	}

	if m.selectedPriority > 0 {
		input.Priority = m.selectedPriority
	}

	if m.selectedAssignee >= 0 && m.selectedAssignee < len(m.users) {
		input.AssigneeID = m.users[m.selectedAssignee].ID
	}

	return input
}

// View renders the create form
func (m CreateModel) View() string {
	// Header
	header := theme.TitleStyle.Render("Create Issue")

	// Form fields
	var fields []string

	// Title
	titleLabel := m.fieldLabel("Title", fieldTitle)
	titleStyle := theme.InputStyle
	if m.focusIndex == fieldTitle {
		titleStyle = theme.InputFocusedStyle
	}
	titleField := titleStyle.Render(m.titleInput.View())
	fields = append(fields, titleLabel+"\n"+titleField)

	// Description
	descLabel := m.fieldLabel("Description", fieldDescription)
	descStyle := theme.InputStyle
	if m.focusIndex == fieldDescription {
		descStyle = theme.InputFocusedStyle
	}
	descField := descStyle.Render(m.descInput.View())
	fields = append(fields, descLabel+"\n"+descField)

	// Team
	teamLabel := m.fieldLabel("Team", fieldTeam)
	teamValue := "None"
	if m.selectedTeam >= 0 && m.selectedTeam < len(m.teams) {
		teamValue = m.teams[m.selectedTeam].Name
	}
	teamField := m.selectField(teamValue, m.focusIndex == fieldTeam)
	fields = append(fields, teamLabel+"  "+teamField)

	// Project
	projectLabel := m.fieldLabel("Project", fieldProject)
	projectValue := "None"
	if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		projectValue = m.projects[m.selectedProject].Name
	}
	projectField := m.selectField(projectValue, m.focusIndex == fieldProject)
	fields = append(fields, projectLabel+"  "+projectField)

	// Priority
	priorityLabel := m.fieldLabel("Priority", fieldPriority)
	priorityValue := theme.PriorityIcon(m.selectedPriority) + " " + theme.PriorityLabel(m.selectedPriority)
	priorityField := m.selectField(priorityValue, m.focusIndex == fieldPriority)
	fields = append(fields, priorityLabel+"  "+priorityField)

	// Assignee
	assigneeLabel := m.fieldLabel("Assignee", fieldAssignee)
	assigneeValue := "Unassigned"
	if m.selectedAssignee >= 0 && m.selectedAssignee < len(m.users) {
		assigneeValue = m.users[m.selectedAssignee].Name
	}
	assigneeField := m.selectField(assigneeValue, m.focusIndex == fieldAssignee)
	fields = append(fields, assigneeLabel+"  "+assigneeField)

	// Help
	help := theme.HelpStyle.Render("Tab: next field  ←/→: change selection  Ctrl+S: submit  Esc: cancel")

	// Combine
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinVertical(lipgloss.Left, fields...),
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(content)
}

// fieldLabel renders a field label
func (m CreateModel) fieldLabel(label string, fieldIndex int) string {
	style := theme.SubtitleStyle
	if m.focusIndex == fieldIndex {
		style = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	}
	return style.Render(label)
}

// selectField renders a select field
func (m CreateModel) selectField(value string, focused bool) string {
	style := theme.ButtonStyle
	if focused {
		style = theme.ButtonActiveStyle
	}
	return style.Render("◄ " + value + " ►")
}

// clamp clamps a value between min and max
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
