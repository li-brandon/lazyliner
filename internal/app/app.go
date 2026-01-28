package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/brandonli/lazyliner/internal/config"
	"github.com/brandonli/lazyliner/internal/git"
	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/components"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/brandonli/lazyliner/internal/ui/views/help"
	"github.com/brandonli/lazyliner/internal/ui/views/issues"
	"github.com/brandonli/lazyliner/internal/ui/views/kanban"
	"github.com/brandonli/lazyliner/internal/ui/views/setup"
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
	ViewEdit
	ViewHelp
	ViewKanban
	ViewSetup
)

// Tab represents the current tab in list view
type Tab int

const (
	TabProject Tab = iota
	TabMyIssues
	TabAllIssues
	TabActive
	TabBacklog
)

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
	searchMode       bool
	searchInput      textinput.Model
	searchQuery      string
	filteredIssues   []linear.Issue
	allProjectIssues []linear.Issue

	// Pagination state
	pageInfo    linear.PageInfo
	loadingMore bool

	// Components
	spinner    spinner.Model
	listView   issues.ListModel
	detailView issues.DetailModel
	createView issues.CreateModel
	editView   issues.EditModel
	helpView   help.Model
	kanbanView kanban.Model
	setupView  setup.Model
	picker     *components.PickerModel
	pickerType string // "status", "assignee", "priority", "project"

	// Current data
	issues         []linear.Issue
	currentIssue   *linear.Issue
	currentProject *linear.Project // Auto-detected from git repo (shows Project tab)
	filterProject  *linear.Project // User-selected project filter (applies to all tabs)
}

func (m Model) tabNames() []string {
	if m.currentProject != nil {
		return []string{"Project", "My Issues", "All Issues", "Active", "Backlog"}
	}
	return []string{"My Issues", "All Issues", "Active", "Backlog"}
}

func (m Model) tabCount() int {
	return len(m.tabNames())
}

func (m Model) tabAtIndex(index int) Tab {
	if m.currentProject != nil {
		tabs := []Tab{TabProject, TabMyIssues, TabAllIssues, TabActive, TabBacklog}
		if index >= 0 && index < len(tabs) {
			return tabs[index]
		}
		return TabProject
	}
	tabs := []Tab{TabMyIssues, TabAllIssues, TabActive, TabBacklog}
	if index >= 0 && index < len(tabs) {
		return tabs[index]
	}
	return TabMyIssues
}

func (m Model) indexOfTab(tab Tab) int {
	if m.currentProject != nil {
		tabs := []Tab{TabProject, TabMyIssues, TabAllIssues, TabActive, TabBacklog}
		for i, t := range tabs {
			if t == tab {
				return i
			}
		}
		return 0
	}
	tabs := []Tab{TabMyIssues, TabAllIssues, TabActive, TabBacklog}
	for i, t := range tabs {
		if t == tab {
			return i
		}
	}
	return 0
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

	// Determine initial view based on API key configuration
	initialView := ViewList
	loading := true
	if cfg.Linear.APIKey == "" {
		initialView = ViewSetup
		loading = false
	}

	return Model{
		config:      cfg,
		keymap:      DefaultKeyMap(),
		client:      linear.NewClient(cfg.Linear.APIKey),
		loading:     loading,
		spinner:     s,
		activeTab:   TabMyIssues,
		view:        initialView,
		searchInput: ti,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	// Don't load data if we're in setup view (no API key)
	if m.view == ViewSetup {
		return nil
	}
	return tea.Batch(
		m.spinner.Tick,
		m.loadInitialData(),
	)
}

// loadInitialData loads the initial data from Linear
func (m Model) loadInitialData() tea.Cmd {
	savedProjectID := m.config.Defaults.Project
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

		var matchedProject *linear.Project

		// First check if there's a saved project filter in config
		if savedProjectID != "" {
			for i := range projects {
				if projects[i].ID == savedProjectID {
					matchedProject = &projects[i]
					break
				}
			}
		}

		// If no saved project, try to match based on repo name
		if matchedProject == nil {
			repoName := git.GetRepoName()
			if repoName != "" {
				repoNameLower := strings.ToLower(repoName)
				repoNameNormalized := strings.ReplaceAll(strings.ReplaceAll(repoNameLower, "-", ""), "_", "")
				for i := range projects {
					projectNameLower := strings.ToLower(projects[i].Name)
					projectNameNormalized := strings.ReplaceAll(strings.ReplaceAll(projectNameLower, "-", ""), "_", "")
					if strings.Contains(projectNameLower, repoNameLower) ||
						strings.Contains(repoNameLower, projectNameLower) ||
						strings.Contains(projectNameNormalized, repoNameNormalized) ||
						strings.Contains(repoNameNormalized, projectNameNormalized) {
						matchedProject = &projects[i]
						break
					}
				}
			}
		}

		return DataLoadedMsg{
			Viewer:         viewer,
			Teams:          teams,
			Projects:       projects,
			MatchedProject: matchedProject,
		}
	}
}

func (m Model) loadIssues() tea.Cmd {
	return m.loadIssuesWithCursor("")
}

func (m Model) loadMoreIssues() tea.Cmd {
	if !m.pageInfo.HasNextPage || m.loadingMore {
		return nil
	}
	return m.loadIssuesWithCursor(m.pageInfo.EndCursor)
}

func (m Model) loadIssuesWithCursor(cursor string) tea.Cmd {
	filterProjectID := ""
	if m.filterProject != nil {
		filterProjectID = m.filterProject.ID
	}
	currentProjectID := ""
	if m.currentProject != nil {
		currentProjectID = m.currentProject.ID
	}
	isAppend := cursor != ""

	return func() tea.Msg {
		ctx := context.Background()
		var conn linear.IssueConnection
		var err error

		switch m.activeTab {
		case TabMyIssues:
			conn, err = m.client.GetMyIssues(ctx, 50, cursor)
			if err == nil && filterProjectID != "" {
				conn.Nodes = filterIssuesByProject(conn.Nodes, filterProjectID)
			}
		case TabAllIssues:
			filter := linear.IssueFilter{Limit: 50, After: cursor}
			if filterProjectID != "" {
				filter.ProjectID = filterProjectID
			}
			conn, err = m.client.GetIssues(ctx, filter)
		case TabActive:
			filter := linear.IssueFilter{
				StateType: "started",
				Limit:     50,
				After:     cursor,
			}
			if filterProjectID != "" {
				filter.ProjectID = filterProjectID
			}
			conn, err = m.client.GetIssues(ctx, filter)
		case TabBacklog:
			filter := linear.IssueFilter{
				StateType: "backlog",
				Limit:     50,
				After:     cursor,
			}
			if filterProjectID != "" {
				filter.ProjectID = filterProjectID
			}
			conn, err = m.client.GetIssues(ctx, filter)
		case TabProject:
			if currentProjectID != "" {
				conn, err = m.client.GetProjectIssues(ctx, currentProjectID, 50, false, cursor)
			}
		}

		return IssuesLoadedMsg{
			Issues:   conn.Nodes,
			PageInfo: conn.PageInfo,
			Append:   isAppend,
			Err:      err,
		}
	}
}

// filterIssuesByProject filters issues to only include those belonging to a specific project
func filterIssuesByProject(issues []linear.Issue, projectID string) []linear.Issue {
	var filtered []linear.Issue
	for _, issue := range issues {
		if issue.Project != nil && issue.Project.ID == projectID {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func (m Model) loadAllProjectIssues() tea.Cmd {
	if m.currentProject == nil {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		conn, err := m.client.GetProjectIssues(ctx, m.currentProject.ID, 100, true, "")
		return AllProjectIssuesLoadedMsg{Issues: conn.Nodes, Err: err}
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

		// Handle picker if it's open
		if m.picker != nil {
			return m.updatePicker(msg)
		}

		// Handle view-specific keys
		switch m.view {
		case ViewList:
			return m.updateListView(msg)
		case ViewDetail:
			return m.updateDetailView(msg)
		case ViewCreate:
			return m.updateCreateView(msg)
		case ViewEdit:
			return m.updateEditView(msg)
		case ViewKanban:
			return m.updateKanbanView(msg)
		case ViewSetup:
			return m.updateSetupView(msg)
		}

	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			if clickedTab := m.getClickedTab(msg.X, msg.Y); clickedTab >= 0 {
				newTab := m.tabAtIndex(clickedTab)
				if newTab != m.activeTab {
					m.activeTab = newTab
					m.loading = true
					return m, m.loadIssues()
				}
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.listView = m.listView.SetSize(msg.Width, msg.Height-4)
		m.detailView = m.detailView.SetSize(msg.Width, msg.Height-4)
		m.createView = m.createView.SetSize(msg.Width, msg.Height-4)
		m.editView = m.editView.SetSize(msg.Width, msg.Height-4)
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
		m.currentProject = msg.MatchedProject
		if m.currentProject != nil {
			m.activeTab = TabProject
		}
		return m, tea.Batch(
			m.loadIssues(),
			m.loadWorkflowStates(),
			m.loadLabels(),
			m.loadUsers(),
		)

	case IssuesLoadedMsg:
		m.loading = false
		m.loadingMore = false
		if msg.Err != nil {
			m.statusMsg = "Error loading issues: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.pageInfo = msg.PageInfo
		if msg.Append {
			m.issues = appendUniqueIssues(m.issues, msg.Issues)
		} else {
			m.issues = msg.Issues
		}
		m.issues = sortIssues(m.issues)
		m.listView = issues.NewListModelWithPagination(m.issues, m.width, m.height-4, m.pageInfo.HasNextPage)
		if msg.PageInfo.HasNextPage && !msg.Append {
			m.statusMsg = fmt.Sprintf("Loaded %d issues (more available, press L)", len(m.issues))
		} else if msg.Append {
			m.statusMsg = fmt.Sprintf("Loaded %d total issues", len(m.issues))
		}
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

	case AllProjectIssuesLoadedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error loading project issues: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.allProjectIssues = sortIssues(msg.Issues)
		m.filterIssues()
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
				// Re-sort issues after update (status/priority may have changed)
				m.issues = sortIssues(m.issues)
				m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
				if m.currentIssue != nil && m.currentIssue.ID == msg.Issue.ID {
					m.currentIssue = msg.Issue
					m.detailView = issues.NewDetailModel(m.currentIssue, m.width, m.height-4)
				}
			}
			if m.view == ViewEdit {
				m.view = ViewDetail
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

	case IssueDeletedMsg:
		if msg.Err != nil {
			m.statusMsg = "Error deleting issue: " + msg.Err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "Issue deleted: " + msg.Identifier
			m.statusErr = false
			m.view = ViewList
			m.currentIssue = nil
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
		if m.activeTab == TabProject && m.currentProject != nil {
			return m, tea.Batch(textinput.Blink, m.loadAllProjectIssues())
		}
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
		currentIndex := m.indexOfTab(m.activeTab)
		nextIndex := (currentIndex + 1) % m.tabCount()
		m.activeTab = m.tabAtIndex(nextIndex)
		m.loading = true
		return m, m.loadIssues()

	case msg.String() == "shift+tab":
		currentIndex := m.indexOfTab(m.activeTab)
		prevIndex := (currentIndex - 1 + m.tabCount()) % m.tabCount()
		m.activeTab = m.tabAtIndex(prevIndex)
		m.loading = true
		return m, m.loadIssues()

	case msg.String() == "1":
		targetTab := TabMyIssues
		if m.currentProject != nil {
			targetTab = TabProject
		}
		if m.activeTab != targetTab {
			m.activeTab = targetTab
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "2":
		targetTab := TabAllIssues
		if m.currentProject != nil {
			targetTab = TabMyIssues
		}
		if m.activeTab != targetTab {
			m.activeTab = targetTab
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "3":
		targetTab := TabActive
		if m.currentProject != nil {
			targetTab = TabAllIssues
		}
		if m.activeTab != targetTab {
			m.activeTab = targetTab
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "4":
		targetTab := TabBacklog
		if m.currentProject != nil {
			targetTab = TabActive
		}
		if m.activeTab != targetTab {
			m.activeTab = targetTab
			m.loading = true
			return m, m.loadIssues()
		}

	case msg.String() == "5":
		if m.currentProject != nil && m.activeTab != TabBacklog {
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
			m.pickerType = "status"
			m.currentIssue = selected
		}
		return m, nil

	case msg.String() == "P":
		// Open project filter picker
		m.picker = components.NewPickerModel("Filter by Project", m.projectsToItems(), m.width, m.height)
		m.pickerType = "project"
		return m, nil

	case msg.String() == "y":
		// Copy branch name
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.copyToClipboard(selected.BranchName, "Branch name copied")
		}

	case msg.String() == "o":
		// Open in Linear (falls back to browser if app not installed)
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.openInLinear(selected.URL)
		}

	case msg.String() == "b":
		m.kanbanView = kanban.New(m.issues, m.states, m.width, m.height-4)
		m.view = ViewKanban
		return m, nil

	case msg.String() == "w":
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.openWorkTask(selected.Identifier)
		}

	case msg.String() == "d":
		if selected := m.listView.SelectedIssue(); selected != nil {
			return m, m.deleteIssue(selected.ID, selected.Identifier)
		}

	case msg.String() == "L":
		if m.pageInfo.HasNextPage && !m.loadingMore {
			m.loadingMore = true
			m.statusMsg = "Loading more issues..."
			return m, m.loadMoreIssues()
		}
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
			m.pickerType = "status"
		}
		return m, nil

	case msg.String() == "a":
		// Open assignee picker
		if m.currentIssue != nil {
			m.picker = components.NewPickerModel("Change Assignee", m.usersToItems(), m.width, m.height)
			m.pickerType = "assignee"
		}
		return m, nil

	case msg.String() == "p":
		// Open priority picker
		if m.currentIssue != nil {
			m.picker = components.NewPickerModel("Change Priority", m.priorityItems(), m.width, m.height)
			m.pickerType = "priority"
		}
		return m, nil

	case msg.String() == "y":
		// Copy branch name
		if m.currentIssue != nil {
			return m, m.copyToClipboard(m.currentIssue.BranchName, "Branch name copied")
		}

	case msg.String() == "o":
		// Open in Linear (falls back to browser if app not installed)
		if m.currentIssue != nil {
			return m, m.openInLinear(m.currentIssue.URL)
		}

	case msg.String() == "w":
		if m.currentIssue != nil {
			return m, m.openWorkTask(m.currentIssue.Identifier)
		}

	case msg.String() == "d":
		if m.currentIssue != nil {
			return m, m.deleteIssue(m.currentIssue.ID, m.currentIssue.Identifier)
		}

	case msg.String() == "e":
		if m.currentIssue != nil {
			m.editView = issues.NewEditModel(m.currentIssue, m.teams, m.projects, m.states, m.users, m.labels, m.width, m.height-4)
			m.view = ViewEdit
		}
		return m, nil
	}

	// Forward to detail view
	var cmd tea.Cmd
	m.detailView, cmd = m.detailView.Update(msg)
	return m, cmd
}

func (m Model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.filteredIssues = nil
		m.allProjectIssues = nil
		m.listView = issues.NewListModel(m.issues, m.width, m.height-4)
		return m, nil

	case "enter":
		m.searchMode = false
		m.searchInput.Blur()
		m.allProjectIssues = nil
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

	searchSource := m.issues
	if m.activeTab == TabProject && len(m.allProjectIssues) > 0 {
		searchSource = m.allProjectIssues
	}

	query := strings.ToLower(m.searchQuery)
	var filtered []linear.Issue
	for _, issue := range searchSource {
		if strings.Contains(strings.ToLower(issue.Title), query) ||
			strings.Contains(strings.ToLower(issue.Identifier), query) ||
			strings.Contains(strings.ToLower(issue.Description), query) {
			filtered = append(filtered, issue)
		}
	}
	m.filteredIssues = sortIssues(filtered)
	m.listView = issues.NewListModel(m.filteredIssues, m.width, m.height-4)
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

func (m Model) updateEditView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.view = ViewDetail
		return m, nil

	case msg.String() == "ctrl+s":
		issueID := m.editView.GetIssueID()
		input := m.editView.GetUpdateInput()
		return m, m.updateIssue(issueID, input)
	}

	var cmd tea.Cmd
	m.editView, cmd = m.editView.Update(msg)
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
			return m, m.openInLinear(selected.URL)
		}

	case "w":
		if selected := m.kanbanView.SelectedIssue(); selected != nil {
			return m, m.openWorkTask(selected.Identifier)
		}

	case "d":
		if selected := m.kanbanView.SelectedIssue(); selected != nil {
			return m, m.deleteIssue(selected.ID, selected.Identifier)
		}
	}

	var cmd tea.Cmd
	m.kanbanView, cmd = m.kanbanView.Update(msg)
	return m, cmd
}

// updatePicker handles picker interactions
func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.picker = nil
		m.pickerType = ""
		return m, nil

	case "enter":
		if m.picker != nil {
			selected := m.picker.SelectedItem()
			if selected != nil {
				return m.handlePickerSelection(selected)
			}
		}
		m.picker = nil
		m.pickerType = ""
		return m, nil
	}

	// Forward navigation keys to picker
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

// handlePickerSelection handles the selection from a picker
func (m Model) handlePickerSelection(item *components.PickerItem) (tea.Model, tea.Cmd) {
	defer func() {
		m.picker = nil
		m.pickerType = ""
	}()

	switch m.pickerType {
	case "status":
		if m.currentIssue != nil {
			return m, m.updateIssueState(m.currentIssue.ID, item.ID)
		}
	case "assignee":
		if m.currentIssue != nil {
			assigneeID := item.ID
			input := linear.IssueUpdateInput{AssigneeID: &assigneeID}
			return m, m.updateIssue(m.currentIssue.ID, input)
		}
	case "priority":
		if m.currentIssue != nil {
			priority := 0
			fmt.Sscanf(item.ID, "%d", &priority)
			input := linear.IssueUpdateInput{Priority: &priority}
			return m, m.updateIssue(m.currentIssue.ID, input)
		}
	case "project":
		// Handle project filter selection
		if item.ID == "" {
			// "All Projects" selected - clear the filter
			m.filterProject = nil
			m.statusMsg = "Showing all projects"
		} else {
			// Find and set the selected project filter
			for i := range m.projects {
				if m.projects[i].ID == item.ID {
					m.filterProject = &m.projects[i]
					m.statusMsg = "Filtering by: " + m.projects[i].Name
					break
				}
			}
		}
		m.statusErr = false
		m.loading = true
		m.picker = nil
		m.pickerType = ""
		return m, m.loadIssues()
	}

	m.picker = nil
	m.pickerType = ""
	return m, nil
}

// createIssue creates a new issue
func (m Model) createIssue(input linear.IssueCreateInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.CreateIssue(ctx, input)
		return IssueCreatedMsg{Issue: issue, Err: err}
	}
}

func (m Model) updateIssue(issueID string, input linear.IssueUpdateInput) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		issue, err := m.client.UpdateIssue(ctx, issueID, input)
		return IssueUpdatedMsg{Issue: issue, Err: err}
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

func (m Model) deleteIssue(issueID, identifier string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.client.DeleteIssue(ctx, issueID)
		return IssueDeletedMsg{IssueID: issueID, Identifier: identifier, Err: err}
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

// openInLinear opens the issue in Linear app if installed, otherwise falls back to browser
func (m Model) openInLinear(url string) tea.Cmd {
	return func() tea.Msg {
		if err := git.OpenInLinear(url); err != nil {
			return StatusMsg{Message: "Failed to open Linear: " + err.Error(), IsError: true}
		}
		return StatusMsg{Message: "Opened in Linear", IsError: false}
	}
}

func (m Model) openWorkTask(identifier string) tea.Cmd {
	return func() tea.Msg {
		workDir, err := os.Getwd()
		if err != nil {
			return StatusMsg{Message: "Failed to get working directory: " + err.Error(), IsError: true}
		}

		cfg := git.TerminalConfig{
			Terminal: m.config.Opencode.Terminal,
			Command:  m.config.Opencode.Command,
		}

		inputCommand := fmt.Sprintf("/work_task %s", identifier)
		if err := git.OpenTerminalWithOpencode(workDir, inputCommand, cfg); err != nil {
			return StatusMsg{Message: "Failed to open terminal: " + err.Error(), IsError: true}
		}
		return StatusMsg{Message: "Opened opencode for " + identifier, IsError: false}
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

// projectsToItems converts projects to picker items
func (m Model) projectsToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.projects)+1)
	// Add "All Projects" option first
	items[0] = components.PickerItem{
		ID:    "",
		Label: "All Projects",
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

// View renders the application
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Show setup view without header/status bar
	if m.view == ViewSetup {
		m.setupView = setup.New(m.width, m.height)
		return m.setupView.View()
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
		case ViewEdit:
			content = m.editView.View()
		case ViewKanban:
			content = m.kanbanView.View()
		}
	}

	// Render status bar
	statusBar := m.renderStatusBar()

	// Combine all parts
	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
	)

	// Overlay picker if open
	if m.picker != nil {
		return m.picker.View()
	}

	return mainView
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := theme.LogoStyle.Render("🦥 Lazyliner")

	var userInfo string
	if m.viewer != nil {
		userInfo = theme.HeaderInfoStyle.Render(m.viewer.Name)
	}

	var tabs string
	for i, name := range m.tabNames() {
		if m.tabAtIndex(i) == m.activeTab {
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
	case ViewDetail:
		keys = []struct {
			key  string
			desc string
		}{
			{"e", "edit"},
			{"s", "status"},
			{"a", "assignee"},
			{"p", "priority"},
			{"y", "copy branch"},
			{"o", "open in linear"},
			{"esc", "back"},
			{"?", "help"},
		}
	case ViewKanban:
		keys = []struct {
			key  string
			desc string
		}{
			{"h/l", "columns"},
			{"j/k", "cards"},
			{"H/L", "move"},
			{"enter", "view"},
			{"d", "delete"},
			{"w", "work"},
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
			{"P", "project"},
			{"b", "board"},
			{"c", "create"},
			{"d", "delete"},
			{"w", "work"},
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

func splitIntoWords(s string) []string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.Fields(s)
}

func appendUniqueIssues(existing, newIssues []linear.Issue) []linear.Issue {
	seen := make(map[string]bool)
	for _, issue := range existing {
		seen[issue.ID] = true
	}
	result := existing
	for _, issue := range newIssues {
		if !seen[issue.ID] {
			result = append(result, issue)
			seen[issue.ID] = true
		}
	}
	return result
}

// stateTypePriority returns the sort priority for a workflow state type.
// Lower values appear first. Incomplete states come before completed ones.
func stateTypePriority(stateType string) int {
	switch stateType {
	case "started":
		return 0 // In progress - highest priority
	case "unstarted":
		return 1 // Not yet started
	case "backlog":
		return 2 // Backlog items
	case "triage":
		return 3 // Triage items
	case "completed":
		return 4 // Done
	case "canceled":
		return 5 // Canceled - lowest priority
	default:
		return 3 // Unknown states go in the middle
	}
}

// sortIssues sorts issues first by completion status (incomplete first),
// then by priority (urgent first, no priority last).
func sortIssues(issuesList []linear.Issue) []linear.Issue {
	sorted := make([]linear.Issue, len(issuesList))
	copy(sorted, issuesList)

	sort.SliceStable(sorted, func(i, j int) bool {
		issueA := sorted[i]
		issueB := sorted[j]

		// Get state types, defaulting to "unstarted" if state is nil
		stateTypeA := "unstarted"
		stateTypeB := "unstarted"
		if issueA.State != nil {
			stateTypeA = issueA.State.Type
		}
		if issueB.State != nil {
			stateTypeB = issueB.State.Type
		}

		// First sort by state type priority (completion status)
		statePriorityA := stateTypePriority(stateTypeA)
		statePriorityB := stateTypePriority(stateTypeB)
		if statePriorityA != statePriorityB {
			return statePriorityA < statePriorityB
		}

		// Then sort by priority (lower number = higher priority)
		// Priority 0 means "no priority" and should come last within the same state
		priorityA := issueA.Priority
		priorityB := issueB.Priority

		// Treat 0 (no priority) as lowest priority (5)
		if priorityA == 0 {
			priorityA = 5
		}
		if priorityB == 0 {
			priorityB = 5
		}

		if priorityA != priorityB {
			return priorityA < priorityB
		}

		// Finally, sort by updated time (most recent first) for stability
		return issueA.UpdatedAt.After(issueB.UpdatedAt)
	})

	return sorted
}

func (m Model) getClickedTab(x, y int) int {
	// Tab bar is on Y=1 (second line, 0-indexed)
	if y != 1 {
		return -1
	}

	// Calculate tab positions: HeaderStyle has Padding(0, 1), so tabs start at X=1
	currentX := 1
	for i, name := range m.tabNames() {
		var renderedTab string
		if m.tabAtIndex(i) == m.activeTab {
			renderedTab = theme.ActiveTabStyle.Render(name)
		} else {
			renderedTab = theme.TabStyle.Render(name)
		}
		tabWidth := lipgloss.Width(renderedTab)
		if x >= currentX && x < currentX+tabWidth {
			return i
		}
		currentX += tabWidth
	}
	return -1
}
