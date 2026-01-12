package linear

import "time"

// Issue represents a Linear issue
type Issue struct {
	ID          string     `json:"id"`
	Identifier  string     `json:"identifier"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    int        `json:"priority"`
	Estimate    *int       `json:"estimate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CanceledAt  *time.Time `json:"canceledAt"`
	DueDate     *string    `json:"dueDate"`
	BranchName  string     `json:"branchName"`
	URL         string     `json:"url"`

	// Relations
	State    *WorkflowState `json:"state"`
	Assignee *User          `json:"assignee"`
	Creator  *User          `json:"creator"`
	Team     *Team          `json:"team"`
	Project  *Project       `json:"project"`
	Cycle    *Cycle         `json:"cycle"`
	Parent   *Issue         `json:"parent"`
	Labels   []Label        `json:"labels"`
}

// WorkflowState represents an issue state
type WorkflowState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Type     string `json:"type"` // backlog, unstarted, started, completed, canceled
	Position int    `json:"position"`
}

// User represents a Linear user
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	AvatarUrl   string `json:"avatarUrl"`
	Active      bool   `json:"active"`
}

// Team represents a Linear team
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// Project represents a Linear project
type Project struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	State       string  `json:"state"` // planned, started, paused, completed, canceled
	Progress    float64 `json:"progress"`
	URL         string  `json:"url"`
}

// Cycle represents a Linear cycle (sprint)
type Cycle struct {
	ID       string    `json:"id"`
	Number   int       `json:"number"`
	Name     string    `json:"name"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
	Progress float64   `json:"progress"`
	IsActive bool      `json:"isActive"`
	IsFuture bool      `json:"isFuture"`
	IsPast   bool      `json:"isPast"`
}

// Label represents an issue label
type Label struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Comment represents an issue comment
type Comment struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	User      *User     `json:"user"`
}

// Viewer represents the currently authenticated user
type Viewer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Active      bool   `json:"active"`
}

// IssueCreateInput represents input for creating an issue
type IssueCreateInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	TeamID      string   `json:"teamId"`
	AssigneeID  string   `json:"assigneeId,omitempty"`
	ProjectID   string   `json:"projectId,omitempty"`
	CycleID     string   `json:"cycleId,omitempty"`
	StateID     string   `json:"stateId,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Estimate    *int     `json:"estimate,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
	ParentID    string   `json:"parentId,omitempty"`
	DueDate     string   `json:"dueDate,omitempty"`
}

// IssueUpdateInput represents input for updating an issue
type IssueUpdateInput struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	AssigneeID  *string  `json:"assigneeId,omitempty"`
	StateID     *string  `json:"stateId,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Estimate    *int     `json:"estimate,omitempty"`
	ProjectID   *string  `json:"projectId,omitempty"`
	CycleID     *string  `json:"cycleId,omitempty"`
	LabelIDs    []string `json:"labelIds,omitempty"`
	ParentID    *string  `json:"parentId,omitempty"`
	DueDate     *string  `json:"dueDate,omitempty"`
}

// IssueFilter represents filters for querying issues
type IssueFilter struct {
	TeamID     string
	ProjectID  string
	AssigneeID string
	StateType  string // backlog, unstarted, started, completed, canceled
	States     []string
	Labels     []string
	Query      string
	Limit      int
	After      string // Cursor for pagination (endCursor from previous page)
}

// Connection types for pagination
type IssueConnection struct {
	Nodes      []Issue  `json:"nodes"`
	PageInfo   PageInfo `json:"pageInfo"`
	TotalCount int      `json:"totalCount,omitempty"`
}

type LabelConnection struct {
	Nodes []Label `json:"nodes"`
}

type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}
