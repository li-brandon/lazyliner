package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/brandonli/lazyliner/internal/config"
	"github.com/brandonli/lazyliner/internal/git"
	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/components"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/brandonli/lazyliner/internal/ui/views/help"
	"github.com/brandonli/lazyliner/internal/ui/views/issues"
	"github.com/brandonli/lazyliner/internal/ui/views/kanban"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents the current view state
type View int

const (
	ViewList View = iota
	ViewDetail
	ViewCreate
	ViewHelp
	ViewKanban
)

// Tab represents the current tab in list view
type Tab int

const (
	TabMyIssues Tab = iota
	TabAllIssues
	TabActive
	TabBacklog
)

var tabNames = []string{"My Issues", "All Issues", "Active", "Backlog"}

// Model is the main application model
type Model struct {
	// Configuration
	config *config.Config
	keymap KeyMap

	// Linear client and data
	client   *linear.Client
	viewer   *linear.Viewer
	teams    []linear.Team
	projects []linear.Project
	users    []linear.User
	states   []linear.WorkflowState
	labels   []linear.Label

	// UI state
	width     int
	height    int
	view      View
	activeTab Tab
	loading   bool
	statusMsg string
	statusErr bool
	showHelp  bool

	// Search state
	searchMode     bool
	searchInput    textinput.Model
	searchQuery    string
	filteredIssues []linear.Issue

	// Components
	spinner    spinner.Model
	listView   issues.ListModel
	detailView issues.DetailModel
	createView issues.CreateModel
	helpView   help.Model
	kanbanView kanban.Model
	picker     *components.PickerModel

	// Current data
	issues       []linear.Issue
	currentIssue *linear.Issue
}

// New creates a new application model
func New(cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Search issues..."
	ti.CharLimit = 100
	ti.Width = 40

	return Model{
		config:      cfg,
		keymap:      DefaultKeyMap(),
		client:      linear.NewClient(cfg.Linear.APIKey),
		loading:     true,
		spinner:     s,
		activeTab:   TabMyIssues,
		view:        ViewList,
		searchInput: ti,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadInitialData(),
	)
}

// loadInitialData loads the initial data from Linear
func (m Model) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		viewer, err := m.client.GetViewer(ctx)
		if err != nil {
			return DataLoadedMsg{Err: err}
		}

		teams, err := m.client.GetTeams(ctx)
		if err != nil {
			return DataLoadedMsg{Err: err}
		}

		projects, err := m.client.GetProjects(ctx)
		if err != nil {
			return DataLoadedMsg{Err: err}
		}

		return DataLoadedMsg{
			Viewer:   viewer,
			Teams:    teams,
			Projects: projects,
		}
	}
}

// loadIssues loads issues based on the current tab
func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var issues []linear.Issue
		var err error

		switch m.activeTab {
		case TabMyIssues:
			issues, err = m.client.GetMyIssues(ctx, 100)
		case TabAllIssues:
			issues, err = m.client.GetIssues(ctx, linear.IssueFilter{Limit: 100})
		case TabActive:
			issues, err = m.client.GetIssues(ctx, linear.IssueFilter{
				StateType: "started",
				Limit:     100,
			})
		case TabBacklog:
			issues, err = m.client.GetIssues(ctx, linear.IssueFilter{
				StateType: "backlog",
				Limit:     100,
			})
		}

		return IssuesLoadedMsg{Issues: issues, Err: err}
	}
}

// loadWorkflowStates loads workflow states for the first team
func (m Model) loadWorkflowStates() tea.Cmd {
	if len(m.teams) == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		states, err := m.client.GetWorkflowStates(ctx, m.teams[0].ID)
		return WorkflowStatesLoadedMsg{States: states, Err: err}
	}
}

// loadLabels loads labels for the first team
func (m Model) loadLabels() tea.Cmd {
	if len(m.teams) == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		labels, err := m.client.GetLabels(ctx, m.teams[0].ID)
		return LabelsLoadedMsg{Labels: labels, Err: err}
	}
}

// loadUsers loads users
func (m Model) loadUsers() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		users, err := m.client.GetUsers(ctx)
		return UsersLoadedMsg{Users: users, Err: err}
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first
		var handled bool
		var cmd tea.Cmd
		m, cmd, handled = m.handleGlobalKeys(msg)
		if handled {
			return m, cmd
		}

		// Handle view-specific keys
		switch m.view {
		case ViewList:
			return m.updateListView(msg)
		case ViewDetail:
			return m.updateDetailView(msg)
		case ViewCreate:
			return m.updateCreateView(msg)
		case ViewKanban:
			return m.updateKanbanView(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.listView = m.listView.SetSize(msg.Width, msg.Height-4)
		m.detailView = m.detailView.SetSize(msg.Width, msg.Height-4)
		m.createView = m.createView.SetSize(msg.Width, msg.Height-4)
		m.kanbanView = m.kanbanView.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case DataLoadedMsg:
		if msg.Err != nil {
			m.loading = false
			m.statusMsg = "Error: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.viewer = msg.Viewer
		m.teams = msg.Teams
		m.projects = msg.Projects
		return m, tea.Batch(
			m.loadIssues(),
			m.loadWorkflowStates(),
			m.loadLabels(),
			m.loadUsers(),
		)

	case IssuesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.statusMsg = "Error loading issues: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.issues = msg.Issues
		m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
		return m, nil

	case WorkflowStatesLoadedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error loading workflow states: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.states = msg.States
		return m, nil

	case LabelsLoadedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error loading labels: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.labels = msg.Labels
		return m, nil

	case UsersLoadedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error loading users: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.users = msg.Users
		return m, nil

	case IssueUpdatedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error: " + msg.Err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "Issue updated"
			m.statusErr = false
			// Update the issue in the list
			if msg.Issue != nil {
				for i, issue := range m.issues {
					if issue.ID == msg.Issue.ID {
						m.issues[i] = *msg.Issue
						break
					}
				}
				m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
				if m.currentIssue != nil && m.currentIssue.ID == msg.Issue.ID {
					m.currentIssue = msg.Issue
					m.detailView = issues.NewDetailModel(m.currentIssue, m.width, m.height-4)
				}
			}
		}
		return m, nil

	case IssueCreatedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error creating issue: " + msg.Err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "Issue created: " + msg.Issue.Identifier
			m.statusErr = false
			m.view = ViewList
			// Refresh issues
			cmds = append(cmds, m.loadIssues())
		}
		return m, tea.Batch(cmds...)

	case StatusMsg:
		m.statusMsg = msg.Message
		m.statusErr = msg.IsError
		return m, nil

	case ClearStatusMsg:
		m.statusMsg = ""
		m.statusErr = false
		return m, nil

	case RefreshMsg:
		m.loading = true
		return m, m.loadIssues()

	case kanban.MoveIssueMsg:
		return m, m.updateIssueState(msg.IssueID, msg.StateID)
	}

	return m, tea.Batch(cmds...)
}

// handleGlobalKeys handles keys that work in all views
func (m Model) handleGlobalKeys(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	switch {
	case msg.String() == "ctrl+c":
		return m, tea.Quit, true
	case msg.String() == "?":
		m.showHelp = !m.showHelp
		return m, nil, true
	case msg.String() == "esc" && m.showHelp:
		m.showHelp = false
		return m, nil, true
	}
	return m, nil, false
}

// updateListView handles updates in the list view
func (m Model) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchMode {
		return m.updateSearchMode(msg)
	}

	switch {
	case msg.String() == "/":
		m.searchMode = true
		m.searchInput.Focus()
		return m, textinput.Blink

	case msg.String() == "q":
		return m, tea.Quit

	case msg.String() == "enter":
		if selected := m.listView.SelectedIssue(); selected != nil {
			m.currentIssue = selected
			m.detailView = issues.NewDetailModel(selected, m.width, m.height-4)
			m.view = ViewDetail
		}
		return m, nil

	case msg.String() == "c":
		m.createView = issues.NewCreateModel(m.teams, m.projects, m.states, m.users, m.labels, m.width, m.height-4)
		m.view = ViewCreate
		return m, nil

	case msg.String() == "tab":
		m.activeTab = Tab((int(m.activeTab) + 1) % len(tabNames))
		m.loading = true
		return m, m.loadIssues()

	case msg.String() == "shift+tab":
		m.activeTab = Tab((int(m.activeTab) - 1 + len(tabNames)) % len(tabNames))
		m.loading = true
		return m, m.loadIssues()

	case msg.String() == "1":
		if m.activeTab != TabMyIssues {
			m.activeTab = TabMyIssues
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "2":
		if m.activeTab != TabAllIssues {
			m.activeTab = TabAllIssues
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "3":
		if m.activeTab != TabActive {
			m.activeTab = TabActive
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "4":
		if m.activeTab != TabBacklog {
			m.activeTab = TabBacklog
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "r":
		m.loading = true
		return m, m.loadIssues()

	case msg.String() == "s":
		// Open status picker
		if selected := m.listView.SelectedIssue(); selected != nil {
			m.picker = components.NewPickerModel("Change Status", m.statesToItems(), m.width, m.height)
			m.currentIssue = selected
		}
		return m, nil

	case msg.String() == "y":
		// Copy branch name
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.copyToClipboard(selected.BranchName, "Branch name copied")
		}

	case msg.String() == "o":
		// Open in browser
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.openInBrowser(selected.URL)
		}

	case msg.String() == "b":
		m.kanbanView = kanban.New(m.issues, m.states, m.width, m.height-4)
		m.view = ViewKanban
		return m, nil
	}

	// Forward to list view
	var cmd tea.Cmd
	m.listView, cmd = m.listView.Update(msg)
	return m, cmd
}

// updateDetailView handles updates in the detail view
func (m Model) updateDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc" || msg.String() == "q":
		m.view = ViewList
		m.currentIssue = nil
		return m, nil

	case msg.String() == "s":
		// Open status picker
		if m.currentIssue != nil {
			m.picker = components.NewPickerModel("Change Status", m.statesToItems(), m.width, m.height)
		}
		return m, nil

	case msg.String() == "a":
		// Open assignee picker
		if m.currentIssue != nil {
			m.picker = components.NewPickerModel("Change Assignee", m.usersToItems(), m.width, m.height)
		}
		return m, nil

	case msg.String() == "p":
		// Open priority picker
		if m.currentIssue != nil {
			m.picker = components.NewPickerModel("Change Priority", m.priorityItems(), m.width, m.height)
		}
		return m, nil

	case msg.String() == "y":
		// Copy branch name
		if m.currentIssue != nil {
			return m, m.copyToClipboard(m.currentIssue.BranchName, "Branch name copied")
		}

	case msg.String() == "o":
		// Open in browser
		if m.currentIssue != nil {
			return m, m.openInBrowser(m.currentIssue.URL)
		}
	}

	// Forward to detail view
	var cmd tea.Cmd
	m.detailView, cmd = m.detailView.Update(msg)
	return m, cmd
}

// updateSearchMode handles updates in search mode
func (m Model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.filteredIssues = nil
		m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
		return m, nil

	case "enter":
		m.searchMode = false
		m.searchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = m.searchInput.Value()
	m.filterIssues()
	return m, cmd
}

// filterIssues filters issues based on search query
func (m *Model) filterIssues() {
	if m.searchQuery == "" {
		m.filteredIssues = nil
		m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
		return
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []linear.Issue
	for _, issue := range m.issues {
		if strings.Contains(strings.ToLower(issue.Title), query) ||
			strings.Contains(strings.ToLower(issue.Identifier), query) ||
			strings.Contains(strings.ToLower(issue.Description), query) {
			filtered = append(filtered, issue)
		}
	}
	m.filteredIssues = filtered
	m.listView = issues.NewListModel(filtered, m.width, m.height-4)
}

// updateCreateView handles updates in the create view
func (m Model) updateCreateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.view = ViewList
		return m, nil

	case msg.String() == "ctrl+s":
		// Submit the form
		input := m.createView.GetInput()
		return m, m.createIssue(input)
	}

	// Forward to create view
	var cmd tea.Cmd
	m.createView, cmd = m.createView.Update(msg)
	return m, cmd
}

func (m Model) updateKanbanView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.view = ViewList
		return m, nil

	case "enter":
		if selected := m.kanbanView.SelectedIssue(); selected != nil {
			m.currentIssue = selected
			m.detailView = issues.NewDetailModel(selected, m.width, m.height-4)
			m.view = ViewDetail
		}
		return m, nil

	case "c":
		m.createView = issues.NewCreateModel(m.teams, m.projects, m.states, m.users, m.labels, m.width, m.height-4)
		m.view = ViewCreate
		return m, nil

	case "r":
		m.loading = true
		m.view = ViewList
		return m, m.loadIssues()

	case "y":
		if selected := m.kanbanView.SelectedIssue(); selected != nil {
			return m, m.copyToClipboard(selected.BranchName, "Branch name copied")
		}

	case "o":
		if selected := m.kanbanView.SelectedIssue(); selected != nil {
			return m, m.openInBrowser(selected.URL)
		}
	}

	var cmd tea.Cmd
	m.kanbanView, cmd = m.kanbanView.Update(msg)
	return m, cmd
}

// createIssue creates a new issue
func (m Model) createIssue(input linear.IssueCreateInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.CreateIssue(ctx, input)
		return IssueCreatedMsg{Issue: issue, Err: err}
	}
}

// updateIssueState updates the state of an issue
func (m Model) updateIssueState(issueID, stateID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.UpdateIssueState(ctx, issueID, stateID)
		return IssueUpdatedMsg{Issue: issue, Err: err}
	}
}

// copyToClipboard copies text to clipboard
func (m Model) copyToClipboard(text, message string) tea.Cmd {
	return func() tea.Msg {
		if err := git.CopyToClipboard(text); err != nil {
			return StatusMsg{Message: "Failed to copy: " + err.Error(), IsError: true}
		}
		return StatusMsg{Message: message, IsError: false}
	}
}

// openInBrowser opens a URL in the default browser
func (m Model) openInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if err := git.OpenInBrowser(url); err != nil {
			return StatusMsg{Message: "Failed to open browser: " + err.Error(), IsError: true}
		}
		return StatusMsg{Message: "Opened in browser", IsError: false}
	}
}

// statesToItems converts workflow states to picker items
func (m Model) statesToItems() []components.PickerItem {
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

// usersToItems converts users to picker items
func (m Model) usersToItems() []components.PickerItem {
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

// priorityItems returns priority picker items
func (m Model) priorityItems() []components.PickerItem {
	return []components.PickerItem{
		{ID: "0", Label: "No Priority", Icon: theme.PriorityIcon(0)},
		{ID: "1", Label: "Urgent", Icon: theme.PriorityIcon(1)},
		{ID: "2", Label: "High", Icon: theme.PriorityIcon(2)},
		{ID: "3", Label: "Medium", Icon: theme.PriorityIcon(3)},
		{ID: "4", Label: "Low", Icon: theme.PriorityIcon(4)},
	}
}

// View renders the application
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if m.showHelp {
		m.helpView = help.New(m.width, m.height)
		return m.helpView.View()
	}

	var content string

	// Render header
	header := m.renderHeader()

	// Render main content based on view
	if m.loading {
		content = m.renderLoading()
	} else {
		switch m.view {
		case ViewList:
			content = m.renderListView()
		case ViewDetail:
			content = m.detailView.View()
		case ViewCreate:
			content = m.createView.View()
		case ViewKanban:
			content = m.kanbanView.View()
		}
	}

	// Render status bar
	statusBar := m.renderStatusBar()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := theme.LogoStyle.Render("🦥 Lazyliner")

	var userInfo string
	if m.viewer != nil {
		userInfo = theme.HeaderInfoStyle.Render(m.viewer.Name)
	}

	// Render tabs
	var tabs string
	for i, name := range tabNames {
		if Tab(i) == m.activeTab {
			tabs += theme.ActiveTabStyle.Render(name)
		} else {
			tabs += theme.TabStyle.Render(name)
		}
	}

	// Build header
	left := title
	right := userInfo
	padding := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if padding < 0 {
		padding = 0
	}

	headerLine := theme.HeaderStyle.Width(m.width).Render(
		left + lipgloss.NewStyle().Width(padding).Render("") + right,
	)

	tabLine := theme.HeaderStyle.Width(m.width).Render(tabs)

	return lipgloss.JoinVertical(lipgloss.Left, headerLine, tabLine)
}

// renderListView renders the issue list view
func (m Model) renderListView() string {
	if m.searchMode || m.searchQuery != "" {
		searchBar := m.renderSearchBar()
		return lipgloss.JoinVertical(lipgloss.Left, searchBar, m.listView.View())
	}
	return m.listView.View()
}

// renderSearchBar renders the search input bar
func (m Model) renderSearchBar() string {
	prefix := theme.TextDimStyle.Render("/ ")
	input := m.searchInput.View()

	count := ""
	if m.searchQuery != "" {
		count = theme.TextDimStyle.Render(fmt.Sprintf(" (%d results)", len(m.filteredIssues)))
	}

	return theme.SearchBarStyle.Width(m.width).Render(prefix + input + count)
}

// renderLoading renders a loading spinner
func (m Model) renderLoading() string {
	return lipgloss.Place(
		m.width,
		m.height-4,
		lipgloss.Center,
		lipgloss.Center,
		m.spinner.View()+" Loading...",
	)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	// Status message
	var status string
	if m.statusMsg != "" {
		if m.statusErr {
			status = theme.ErrorStyle.Render(m.statusMsg)
		} else {
			status = theme.SuccessStyle.Render(m.statusMsg)
		}
	}

	// Help hints
	help := m.renderHelp()

	// Combine
	if status != "" {
		return theme.StatusBarStyle.Width(m.width).Render(
			status + "  " + help,
		)
	}
	return theme.StatusBarStyle.Width(m.width).Render(help)
}

func (m Model) renderHelp() string {
	var keys []struct {
		key  string
		desc string
	}

	switch m.view {
	case ViewKanban:
		keys = []struct {
			key  string
			desc string
		}{
			{"h/l", "columns"},
			{"j/k", "cards"},
			{"H/L", "move"},
			{"enter", "view"},
			{"esc", "list"},
			{"?", "help"},
		}
	default:
		keys = []struct {
			key  string
			desc string
		}{
			{"j/k", "navigate"},
			{"enter", "view"},
			{"/", "search"},
			{"b", "board"},
			{"c", "create"},
			{"?", "help"},
			{"q", "quit"},
		}
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			theme.StatusBarKeyStyle.Render(k.key)+
				theme.StatusBarDescStyle.Render(":"+k.desc),
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, joinWithSep(parts, "  ")...)
}

func joinWithSep(parts []string, sep string) []string {
	if len(parts) == 0 {
		return parts
	}
	result := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		result = append(result, p)
		if i < len(parts)-1 {
			result = append(result, sep)
		}
	}
	return result
}
