package issues

import (
	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/components"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditModel is the issue edit form
type EditModel struct {
	// The issue being edited
	issue *linear.Issue

	// Form fields
	titleInput textinput.Model
	descInput  textarea.Model

	// Options
	teams    []linear.Team
	projects []linear.Project
	states   []linear.WorkflowState
	users    []linear.User
	labels   []linear.Label

	// Selected values (indices)
	selectedTeam     int
	selectedProject  int
	selectedState    int
	selectedPriority int
	selectedAssignee int

	// UI state
	focusIndex int
	width      int
	height     int

	// Picker state
	picker     *components.PickerModel
	pickerType string // "state", "project", "priority", "assignee"
}

// Edit field indices
const (
	editFieldTitle = iota
	editFieldDescription
	editFieldState
	editFieldPriority
	editFieldAssignee
	editFieldProject
	editFieldCount
)

// NewEditModel creates a new edit model pre-populated with issue data
func NewEditModel(issue *linear.Issue, teams []linear.Team, projects []linear.Project, states []linear.WorkflowState, users []linear.User, labels []linear.Label, width, height int) EditModel {
	// Title input
	ti := textinput.New()
	ti.Placeholder = "Issue title"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = width - 20
	ti.SetValue(issue.Title)

	// Description input
	ta := textarea.New()
	ta.Placeholder = "Description (markdown supported)"
	ta.CharLimit = 10000
	ta.SetWidth(width - 20)
	ta.SetHeight(6)
	ta.SetValue(issue.Description)

	// Find selected indices based on current issue values
	selectedTeam := 0
	if issue.Team != nil {
		for i, t := range teams {
			if t.ID == issue.Team.ID {
				selectedTeam = i
				break
			}
		}
	}

	selectedProject := -1 // -1 means no project
	if issue.Project != nil {
		for i, p := range projects {
			if p.ID == issue.Project.ID {
				selectedProject = i
				break
			}
		}
	}

	selectedState := 0
	if issue.State != nil {
		for i, s := range states {
			if s.ID == issue.State.ID {
				selectedState = i
				break
			}
		}
	}

	selectedAssignee := -1 // -1 means unassigned
	if issue.Assignee != nil {
		for i, u := range users {
			if u.ID == issue.Assignee.ID {
				selectedAssignee = i
				break
			}
		}
	}

	return EditModel{
		issue:            issue,
		titleInput:       ti,
		descInput:        ta,
		teams:            teams,
		projects:         projects,
		states:           states,
		users:            users,
		labels:           labels,
		selectedTeam:     selectedTeam,
		selectedProject:  selectedProject,
		selectedState:    selectedState,
		selectedPriority: issue.Priority,
		selectedAssignee: selectedAssignee,
		focusIndex:       editFieldTitle,
		width:            width,
		height:           height,
	}
}

// SetSize updates the form dimensions
func (m EditModel) SetSize(width, height int) EditModel {
	m.width = width
	m.height = height
	if m.titleInput.Placeholder != "" {
		m.titleInput.Width = width - 20
	}
	if m.descInput.Placeholder != "" {
		m.descInput.SetWidth(width - 20)
	}
	return m
}

// Update handles messages
func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle picker if open
		if m.picker != nil {
			return m.updatePicker(msg)
		}

		switch msg.String() {
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % editFieldCount
			m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + editFieldCount) % editFieldCount
			m.updateFocus()
		case "left":
			m.handleLeftRight(-1)
		case "right":
			m.handleLeftRight(1)
		case "enter":
			// Open picker for select fields
			m.openPickerForField()
		default:
			// Forward to focused field
			switch m.focusIndex {
			case editFieldTitle:
				var cmd tea.Cmd
				m.titleInput, cmd = m.titleInput.Update(msg)
				cmds = append(cmds, cmd)
			case editFieldDescription:
				var cmd tea.Cmd
				m.descInput, cmd = m.descInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// openPickerForField opens the appropriate picker based on the focused field
func (m *EditModel) openPickerForField() {
	switch m.focusIndex {
	case editFieldState:
		m.picker = components.NewPickerModel("Select Status", m.statesToItems(), m.width, m.height)
		m.pickerType = "state"
	case editFieldPriority:
		m.picker = components.NewPickerModel("Select Priority", m.priorityItems(), m.width, m.height)
		m.pickerType = "priority"
	case editFieldAssignee:
		m.picker = components.NewPickerModel("Select Assignee", m.usersToItems(), m.width, m.height)
		m.pickerType = "assignee"
	case editFieldProject:
		m.picker = components.NewPickerModel("Select Project", m.projectsToItems(), m.width, m.height)
		m.pickerType = "project"
	}
}

// updatePicker handles picker interactions
func (m EditModel) updatePicker(msg tea.KeyMsg) (EditModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.picker = nil
		m.pickerType = ""
		return m, nil

	case "enter":
		if m.picker != nil {
			selected := m.picker.SelectedItem()
			if selected != nil {
				m.handlePickerSelection(selected)
			}
		}
		m.picker = nil
		m.pickerType = ""
		return m, nil
	}

	// Forward to picker
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

// handlePickerSelection handles the selection from a picker
func (m *EditModel) handlePickerSelection(item *components.PickerItem) {
	switch m.pickerType {
	case "state":
		for i, state := range m.states {
			if state.ID == item.ID {
				m.selectedState = i
				break
			}
		}
	case "project":
		if item.ID == "" {
			m.selectedProject = -1
		} else {
			for i, project := range m.projects {
				if project.ID == item.ID {
					m.selectedProject = i
					break
				}
			}
		}
	case "priority":
		var priority int
		switch item.ID {
		case "0":
			priority = 0
		case "1":
			priority = 1
		case "2":
			priority = 2
		case "3":
			priority = 3
		case "4":
			priority = 4
		}
		m.selectedPriority = priority
	case "assignee":
		if item.ID == "" {
			m.selectedAssignee = -1
		} else {
			for i, user := range m.users {
				if user.ID == item.ID {
					m.selectedAssignee = i
					break
				}
			}
		}
	}
}

// statesToItems converts workflow states to picker items
func (m EditModel) statesToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.states))
	for i, s := range m.states {
		items[i] = components.PickerItem{
			ID:    s.ID,
			Label: s.Name,
			Icon:  theme.StatusIcon(s.Type),
		}
	}
	return items
}

// projectsToItems converts projects to picker items
func (m EditModel) projectsToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.projects)+1)
	items[0] = components.PickerItem{
		ID:    "",
		Label: "None",
		Icon:  "📁",
	}
	for i, p := range m.projects {
		icon := "📁"
		if p.Icon != "" {
			icon = p.Icon
		}
		items[i+1] = components.PickerItem{
			ID:    p.ID,
			Label: p.Name,
			Icon:  icon,
		}
	}
	return items
}

// priorityItems returns priority picker items
func (m EditModel) priorityItems() []components.PickerItem {
	return []components.PickerItem{
		{ID: "0", Label: "No Priority", Icon: theme.PriorityIcon(0)},
		{ID: "1", Label: "Urgent", Icon: theme.PriorityIcon(1)},
		{ID: "2", Label: "High", Icon: theme.PriorityIcon(2)},
		{ID: "3", Label: "Medium", Icon: theme.PriorityIcon(3)},
		{ID: "4", Label: "Low", Icon: theme.PriorityIcon(4)},
	}
}

// usersToItems converts users to picker items
func (m EditModel) usersToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.users)+1)
	items[0] = components.PickerItem{
		ID:    "",
		Label: "Unassigned",
		Icon:  "👤",
	}
	for i, u := range m.users {
		items[i+1] = components.PickerItem{
			ID:    u.ID,
			Label: u.Name,
			Icon:  "👤",
		}
	}
	return items
}

// updateFocus updates which field is focused
func (m *EditModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()

	switch m.focusIndex {
	case editFieldTitle:
		m.titleInput.Focus()
	case editFieldDescription:
		m.descInput.Focus()
	}
}

// handleLeftRight handles left/right navigation for select fields
func (m *EditModel) handleLeftRight(dir int) {
	switch m.focusIndex {
	case editFieldState:
		m.selectedState = clamp(m.selectedState+dir, 0, len(m.states)-1)
	case editFieldPriority:
		m.selectedPriority = clamp(m.selectedPriority+dir, 0, 4)
	case editFieldAssignee:
		m.selectedAssignee = clamp(m.selectedAssignee+dir, -1, len(m.users)-1)
	case editFieldProject:
		m.selectedProject = clamp(m.selectedProject+dir, -1, len(m.projects)-1)
	}
}

// GetIssueID returns the ID of the issue being edited
func (m EditModel) GetIssueID() string {
	if m.issue == nil {
		return ""
	}
	return m.issue.ID
}

// GetUpdateInput returns the current form input as IssueUpdateInput
func (m EditModel) GetUpdateInput() linear.IssueUpdateInput {
	title := m.titleInput.Value()
	description := m.descInput.Value()

	input := linear.IssueUpdateInput{
		Title:       &title,
		Description: &description,
		Priority:    &m.selectedPriority,
	}

	// State
	if m.selectedState >= 0 && m.selectedState < len(m.states) {
		stateID := m.states[m.selectedState].ID
		input.StateID = &stateID
	}

	// Project (can be nil to unset). A selectedProject of -1 means "None".
	if m.selectedProject == -1 {
		// Explicitly unset the project on the issue.
		input.ProjectID = nil
	} else if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		projectID := m.projects[m.selectedProject].ID
		input.ProjectID = &projectID
	}

	// Assignee (can be nil to unassign)
	if m.selectedAssignee == -1 {
		// Explicitly set to nil to unassign in Linear
		input.AssigneeID = nil
	} else if m.selectedAssignee >= 0 && m.selectedAssignee < len(m.users) {
		assigneeID := m.users[m.selectedAssignee].ID
		input.AssigneeID = &assigneeID
	}

	return input
}

// View renders the edit form
func (m EditModel) View() string {
	// If picker is open, render it instead
	if m.picker != nil {
		return m.picker.View()
	}

	// Header
	headerText := "Edit Issue"
	if m.issue != nil {
		headerText = "Edit Issue: " + m.issue.Identifier
	}
	header := theme.TitleStyle.Render(headerText)

	// Form fields
	var fields []string

	// Title
	titleLabel := m.fieldLabel("Title", editFieldTitle)
	titleStyle := theme.InputStyle
	if m.focusIndex == editFieldTitle {
		titleStyle = theme.InputFocusedStyle
	}
	titleField := titleStyle.Render(m.titleInput.View())
	fields = append(fields, titleLabel+"\n"+titleField)

	// Description
	descLabel := m.fieldLabel("Description", editFieldDescription)
	descStyle := theme.InputStyle
	if m.focusIndex == editFieldDescription {
		descStyle = theme.InputFocusedStyle
	}
	descField := descStyle.Render(m.descInput.View())
	fields = append(fields, descLabel+"\n"+descField)

	// State
	stateLabel := m.fieldLabel("Status", editFieldState)
	stateValue := "None"
	if m.selectedState >= 0 && m.selectedState < len(m.states) {
		stateValue = theme.StatusIcon(m.states[m.selectedState].Type) + " " + m.states[m.selectedState].Name
	}
	stateField := m.selectField(stateValue, m.focusIndex == editFieldState)
	fields = append(fields, stateLabel+"  "+stateField)

	// Priority
	priorityLabel := m.fieldLabel("Priority", editFieldPriority)
	priorityValue := theme.PriorityIcon(m.selectedPriority) + " " + theme.PriorityLabel(m.selectedPriority)
	priorityField := m.selectField(priorityValue, m.focusIndex == editFieldPriority)
	fields = append(fields, priorityLabel+"  "+priorityField)

	// Assignee
	assigneeLabel := m.fieldLabel("Assignee", editFieldAssignee)
	assigneeValue := "Unassigned"
	if m.selectedAssignee >= 0 && m.selectedAssignee < len(m.users) {
		assigneeValue = m.users[m.selectedAssignee].Name
	}
	assigneeField := m.selectField(assigneeValue, m.focusIndex == editFieldAssignee)
	fields = append(fields, assigneeLabel+"  "+assigneeField)

	// Project
	projectLabel := m.fieldLabel("Project", editFieldProject)
	projectValue := "None"
	if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		projectValue = m.projects[m.selectedProject].Name
	}
	projectField := m.selectField(projectValue, m.focusIndex == editFieldProject)
	fields = append(fields, projectLabel+"  "+projectField)

	// Help
	help := theme.HelpStyle.Render("Tab: next  Enter: select  ←/→: quick change  Ctrl+S: save  Esc: cancel")

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
func (m EditModel) fieldLabel(label string, fieldIndex int) string {
	style := theme.SubtitleStyle
	if m.focusIndex == fieldIndex {
		style = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	}
	return style.Render(label)
}

// selectField renders a select field
func (m EditModel) selectField(value string, focused bool) string {
	style := theme.ButtonStyle
	if focused {
		style = theme.ButtonActiveStyle
	}
	return style.Render("◄ " + value + " ►")
}
