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

	// Components
	spinner    spinner.Model
	listView   issues.ListModel
	detailView issues.DetailModel
	createView issues.CreateModel
	editView   issues.EditModel
	helpView   help.Model
	kanbanView kanban.Model
	picker     *components.PickerModel

	// Current data
	issues         []linear.Issue
	currentIssue   *linear.Issue
	currentProject *linear.Project
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

		var matchedProject *linear.Project
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

		return DataLoadedMsg{
			Viewer:         viewer,
			Teams:          teams,
			Projects:       projects,
			MatchedProject: matchedProject,
		}
	}
}

// loadIssues loads issues based on the current tab
func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var loadedIssues []linear.Issue
		var err error

		switch m.activeTab {
		case TabMyIssues:
			loadedIssues, err = m.client.GetMyIssues(ctx, 100)
		case TabAllIssues:
			loadedIssues, err = m.client.GetIssues(ctx, linear.IssueFilter{Limit: 100})
		case TabActive:
			loadedIssues, err = m.client.GetIssues(ctx, linear.IssueFilter{
				StateType: "started",
				Limit:     100,
			})
		case TabBacklog:
			loadedIssues, err = m.client.GetIssues(ctx, linear.IssueFilter{
				StateType: "backlog",
				Limit:     100,
			})
		case TabProject:
			if m.currentProject != nil {
				loadedIssues, err = m.client.GetProjectIssues(ctx, m.currentProject.ID, 100, false)
			}
		}

		return IssuesLoadedMsg{Issues: loadedIssues, Err: err}
	}
}

func (m Model) loadAllProjectIssues() tea.Cmd {
	if m.currentProject == nil {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		allIssues, err := m.client.GetProjectIssues(ctx, m.currentProject.ID, 100, true)
		return AllProjectIssuesLoadedMsg{Issues: allIssues, Err: err}
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
		case ViewEdit:
			return m.updateEditView(msg)
		case ViewKanban:
			return m.updateKanbanView(msg)
		}

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
		if msg.Err != nil {
			m.statusMsg = "Error loading issues: " + msg.Err.Error()
			m.statusErr = true
			return m, nil
		}
		m.issues = sortIssues(msg.Issues)
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
			m.currentIssue = selected
		}
		return m, nil

	case msg.String() == "a":
		// Open assignee picker
		if selected := m.listView.SelectedIssue(); selected != nil {
			m.picker = components.NewPickerModel("Change Assignee", m.usersToItems(), m.width, m.height)
			m.currentIssue = selected
		}
		return m, nil

	case msg.String() == "p":
		// Open priority picker
		if selected := m.listView.SelectedIssue(); selected != nil {
			m.picker = components.NewPickerModel("Change Priority", m.priorityItems(), m.width, m.height)
			m.currentIssue = selected
		}
		return m, nil

	case msg.String() == "l":
		// Open labels picker
		if selected := m.listView.SelectedIssue(); selected != nil {
			m.picker = components.NewPickerModel("Manage Labels", m.labelsToItems(), m.width, m.height)
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

	case msg.String() == "O":
		// Open in Linear
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

	case msg.String() == "O":
		// Open in Linear
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
			return m, m.openInBrowser(selected.URL)
		}

	case "O":
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

// openInBrowser opens a URL in the default browser
func (m Model) openInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if err := git.OpenInBrowser(url); err != nil {
			return StatusMsg{Message: "Failed to open browser: " + err.Error(), IsError: true}
		}
		return StatusMsg{Message: "Opened in browser", IsError: false}
	}
}

// openInLinear opens the issue URL in the Linear desktop app
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

// labelsToItems converts labels to picker items
func (m Model) labelsToItems() []components.PickerItem {
	items := make([]components.PickerItem, len(m.labels))
	for i, l := range m.labels {
		items[i] = components.PickerItem{
			ID:    l.ID,
			Label: l.Name,
			Icon:  "🏷️",
		}
	}
	return items
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
		case ViewEdit:
			content = m.editView.View()
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
			{"o/O", "open"},
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
