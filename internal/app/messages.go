package app

import "github.com/brandonli/lazyliner/internal/linear"

// Message types for the application

// DataLoadedMsg is sent when initial data is loaded
type DataLoadedMsg struct {
	Viewer             *linear.Viewer
	Teams              []linear.Team
	Projects           []linear.Project
	MatchedProject     *linear.Project // Auto-detected from git repo (for Project tab)
	SavedFilterProject *linear.Project // Restored from config (for project filter)
	Err                error
}

// IssuesLoadedMsg is sent when issues are loaded
type IssuesLoadedMsg struct {
	Issues []linear.Issue
	Err    error
}

// IssueLoadedMsg is sent when a single issue is loaded
type IssueLoadedMsg struct {
	Issue *linear.Issue
	Err   error
}

// IssueCreatedMsg is sent when an issue is created
type IssueCreatedMsg struct {
	Issue *linear.Issue
	Err   error
}

// IssueUpdatedMsg is sent when an issue is updated
type IssueUpdatedMsg struct {
	Issue *linear.Issue
	Err   error
}

// IssueDeletedMsg is sent when an issue is deleted
type IssueDeletedMsg struct {
	IssueID    string
	Identifier string
	Err        error
}

// WorkflowStatesLoadedMsg is sent when workflow states are loaded
type WorkflowStatesLoadedMsg struct {
	States []linear.WorkflowState
	Err    error
}

// LabelsLoadedMsg is sent when labels are loaded
type LabelsLoadedMsg struct {
	Labels []linear.Label
	Err    error
}

// UsersLoadedMsg is sent when users are loaded
type UsersLoadedMsg struct {
	Users []linear.User
	Err   error
}

// SearchResultsMsg is sent when search results are returned
type SearchResultsMsg struct {
	Issues []linear.Issue
	Query  string
	Err    error
}

// ErrorMsg represents a generic error
type ErrorMsg struct {
	Err     error
	Context string
}

// StatusMsg is a temporary status message to display
type StatusMsg struct {
	Message string
	IsError bool
}

// ClearStatusMsg clears the status message
type ClearStatusMsg struct{}

// RefreshMsg triggers a data refresh
type RefreshMsg struct{}

// SwitchTabMsg switches to a different tab
type SwitchTabMsg struct {
	Tab int
}

// OpenIssueMsg opens an issue in detail view
type OpenIssueMsg struct {
	Issue *linear.Issue
}

// CloseDetailMsg closes the detail view
type CloseDetailMsg struct{}

// OpenCreateMsg opens the create issue form
type OpenCreateMsg struct{}

// CloseCreateMsg closes the create issue form
type CloseCreateMsg struct{}

// OpenPickerMsg opens a picker modal
type OpenPickerMsg struct {
	Type string // "status", "assignee", "priority", "labels", "project"
}

// ClosePickerMsg closes the picker modal
type ClosePickerMsg struct{}

// CopyToClipboardMsg copies text to clipboard
type CopyToClipboardMsg struct {
	Text    string
	Message string // Success message to show
}

// OpenInBrowserMsg opens a URL in the browser
type OpenInBrowserMsg struct {
	URL string
}

// AllProjectIssuesLoadedMsg is sent when all project issues (including completed) are loaded for search
type AllProjectIssuesLoadedMsg struct {
	Issues []linear.Issue
	Err    error
}
