package issues

import (
	"strings"

	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/components"
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
	focusIndex   int
	scrollOffset int
	width        int
	height       int

	// Picker state
	picker     *components.PickerModel
	pickerType string // "team", "project", "priority", "assignee"
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
		// Handle picker if open
		if m.picker != nil {
			return m.updatePicker(msg)
		}

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
		case "enter":
			// Open picker for select fields
			m.openPickerForField()
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

// openPickerForField opens the appropriate picker based on the focused field
func (m *CreateModel) openPickerForField() {
	switch m.focusIndex {
	case fieldTeam:
		m.picker = components.NewPickerModel("Select Team", m.teamsToItems(), m.width, m.height)
		m.pickerType = "team"
	case fieldProject:
		m.picker = components.NewPickerModel("Select Project", m.projectsToItems(), m.width, m.height)
		m.pickerType = "project"
	case fieldPriority:
		m.picker = components.NewPickerModel("Select Priority", m.priorityItems(), m.width, m.height)
		m.pickerType = "priority"
	case fieldAssignee:
		m.picker = components.NewPickerModel("Select Assignee", m.usersToItems(), m.width, m.height)
		m.pickerType = "assignee"
	}
}

// updatePicker handles picker interactions
func (m CreateModel) updatePicker(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
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
func (m *CreateModel) handlePickerSelection(item *components.PickerItem) {
	switch m.pickerType {
	case "team":
		for i, team := range m.teams {
			if team.ID == item.ID {
				m.selectedTeam = i
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

// teamsToItems converts teams to picker items
func (m CreateModel) teamsToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.teams))
	for i, t := range m.teams {
		items[i] = components.PickerItem{
			ID:    t.ID,
			Label: t.Name,
			Icon:  "👥",
		}
	}
	return items
}

// projectsToItems converts projects to picker items
func (m CreateModel) projectsToItems() []components.PickerItem {
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
func (m CreateModel) priorityItems() []components.PickerItem {
	return []components.PickerItem{
		{ID: "0", Label: "No Priority", Icon: theme.PriorityIcon(0)},
		{ID: "1", Label: "Urgent", Icon: theme.PriorityIcon(1)},
		{ID: "2", Label: "High", Icon: theme.PriorityIcon(2)},
		{ID: "3", Label: "Medium", Icon: theme.PriorityIcon(3)},
		{ID: "4", Label: "Low", Icon: theme.PriorityIcon(4)},
	}
}

// usersToItems converts users to picker items
func (m CreateModel) usersToItems() []components.PickerItem {
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
func (m *CreateModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()

	switch m.focusIndex {
	case fieldTitle:
		m.titleInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	}

	m.ensureFocusVisible()
}

func (m *CreateModel) fieldHeights() []int {
	return []int{4, 9, 2, 2, 2, 2}
}

func (m *CreateModel) ensureFocusVisible() {
	heights := m.fieldHeights()

	fieldTop := 3
	for i := 0; i < m.focusIndex; i++ {
		fieldTop += heights[i]
	}
	fieldBottom := fieldTop + heights[m.focusIndex]

	viewportHeight := m.height - 4

	if fieldTop < m.scrollOffset {
		m.scrollOffset = fieldTop
	}

	if fieldBottom > m.scrollOffset+viewportHeight {
		m.scrollOffset = fieldBottom - viewportHeight
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
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
	// If picker is open, render it instead
	if m.picker != nil {
		return m.picker.View()
	}

	header := theme.TitleStyle.Render("Create Issue")

	var fields []string

	titleLabel := m.fieldLabel("Title", fieldTitle)
	titleStyle := theme.InputStyle
	if m.focusIndex == fieldTitle {
		titleStyle = theme.InputFocusedStyle
	}
	titleField := titleStyle.Render(m.titleInput.View())
	fields = append(fields, titleLabel+"\n"+titleField)

	descLabel := m.fieldLabel("Description", fieldDescription)
	descStyle := theme.InputStyle
	if m.focusIndex == fieldDescription {
		descStyle = theme.InputFocusedStyle
	}
	descField := descStyle.Render(m.descInput.View())
	fields = append(fields, descLabel+"\n"+descField)

	teamLabel := m.fieldLabel("Team", fieldTeam)
	teamValue := "None"
	if m.selectedTeam >= 0 && m.selectedTeam < len(m.teams) {
		teamValue = m.teams[m.selectedTeam].Name
	}
	teamField := m.selectField(teamValue, m.focusIndex == fieldTeam)
	fields = append(fields, teamLabel+"  "+teamField)

	projectLabel := m.fieldLabel("Project", fieldProject)
	projectValue := "None"
	if m.selectedProject >= 0 && m.selectedProject < len(m.projects) {
		projectValue = m.projects[m.selectedProject].Name
	}
	projectField := m.selectField(projectValue, m.focusIndex == fieldProject)
	fields = append(fields, projectLabel+"  "+projectField)

	priorityLabel := m.fieldLabel("Priority", fieldPriority)
	priorityValue := theme.PriorityIcon(m.selectedPriority) + " " + theme.PriorityLabel(m.selectedPriority)
	priorityField := m.selectField(priorityValue, m.focusIndex == fieldPriority)
	fields = append(fields, priorityLabel+"  "+priorityField)

	assigneeLabel := m.fieldLabel("Assignee", fieldAssignee)
	assigneeValue := "Unassigned"
	if m.selectedAssignee >= 0 && m.selectedAssignee < len(m.users) {
		assigneeValue = m.users[m.selectedAssignee].Name
	}
	assigneeField := m.selectField(assigneeValue, m.focusIndex == fieldAssignee)
	fields = append(fields, assigneeLabel+"  "+assigneeField)

	help := theme.HelpStyle.Render("Tab: next  Enter: select  ←/→: quick change  Ctrl+S: submit  Esc: cancel")

	formContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinVertical(lipgloss.Left, fields...),
		"",
		help,
	)

	lines := strings.Split(formContent, "\n")
	viewportHeight := m.height - 2

	if len(lines) <= viewportHeight || m.scrollOffset == 0 && len(lines) <= viewportHeight {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Width(m.width).
			Height(m.height).
			Render(formContent)
	}

	startLine := m.scrollOffset
	if startLine > len(lines)-viewportHeight {
		startLine = len(lines) - viewportHeight
	}
	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + viewportHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleContent := strings.Join(lines[startLine:endLine], "\n")

	scrollIndicator := ""
	if startLine > 0 {
		scrollIndicator = "▲ "
	}
	if endLine < len(lines) {
		if scrollIndicator != "" {
			scrollIndicator += "▼"
		} else {
			scrollIndicator = "▼"
		}
	}

	if scrollIndicator != "" {
		visibleContent = visibleContent + "\n" + theme.TextMutedStyle.Render(scrollIndicator)
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height).
		Render(visibleContent)
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
