# AGENTS.md - Lazyliner

A terminal TUI for Linear built with Go and Bubble Tea.

## Quick Reference

```bash
make build          # Build binary to bin/lazyliner
make run            # Run in development mode
make test           # Run all tests
make lint           # Run golangci-lint
make fmt            # Format with gofmt -s
make tidy           # go mod tidy

# Single test
go test -v ./internal/linear -run TestGetIssues
```

## Project Structure

```
cmd/lazyliner/         # CLI entrypoint (Cobra commands)
internal/
  app/                 # Main Bubble Tea model, keymaps, messages
  config/              # Viper config management
  git/                 # Git/clipboard/browser utilities
  linear/              # Linear GraphQL API client
  ui/
    components/        # Reusable UI components (picker)
    theme/             # Colors and Lipgloss styles
    views/             # View models (issues list/detail/create, help)
```

## Code Style

### Imports

```go
import (
    "context"                                          // stdlib first
    "fmt"

    "github.com/brandonli/lazyliner/internal/config"   // internal second
    "github.com/brandonli/lazyliner/internal/linear"

    tea "github.com/charmbracelet/bubbletea"           // external third, alias tea
    "github.com/charmbracelet/lipgloss"
)
```

### Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Packages | lowercase, single word | `linear`, `theme` |
| Types/Structs | PascalCase | `IssueCreateInput` |
| Exported funcs | PascalCase | `GetIssues()` |
| Private funcs | camelCase | `buildFilter()` |
| JSON tags | snake_case | `json:"branch_name"` |
| Config tags | snake_case | `mapstructure:"api_key"` |

### Enums with iota

```go
type View int
const (
    ViewList View = iota
    ViewDetail
    ViewCreate
)
```

### Receivers

- **Value receivers**: read-only methods returning new values
- **Pointer receivers**: methods that modify state

### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to fetch issues: %w", err)
}

// Always check nullable fields
status := "Unknown"
if issue.State != nil {
    status = issue.State.Name
}
```

## Bubble Tea Patterns

### Message Types (in messages.go)

```go
type IssuesLoadedMsg struct {
    Issues []linear.Issue
    Err    error
}
```

### Commands (Async Operations)

```go
func (m Model) loadIssues() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        issues, err := m.client.GetIssues(ctx, filter)
        return IssuesLoadedMsg{Issues: issues, Err: err}
    }
}
```

### Update Pattern

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keyboard input
    case IssuesLoadedMsg:
        if msg.Err != nil {
            m.statusMsg = "Error: " + msg.Err.Error()
            return m, nil
        }
        m.issues = msg.Issues
    }
    return m, nil
}
```

## Linear API Client

GraphQL queries go in `queries.go`, mutations in `mutations.go`:

```go
func (c *Client) GetTeams(ctx context.Context) ([]Team, error) {
    query := `query Teams { teams { nodes { id name key } } }`
    var result struct {
        Teams struct { Nodes []Team `json:"nodes"` } `json:"teams"`
    }
    if err := c.execute(ctx, query, nil, &result); err != nil {
        return nil, err
    }
    return result.Teams.Nodes, nil
}
```

## UI/Theme

Styles in `theme/styles.go`, colors in `theme/colors.go`:

```go
var ListItemSelectedStyle = lipgloss.NewStyle().
    Foreground(TextBright).
    Background(SurfaceHover).
    Padding(0, 1)
```

Key colors: `Primary` (#5E6AD2), `Success`, `Warning`, `Danger`, `Text`, `TextMuted`

## Common Tasks

### Adding a New View
1. Create view model in `internal/ui/views/`
2. Add `View` enum value in `internal/app/app.go`
3. Add case in `Update()` and `View()` switch statements
4. Add navigation keybinding in `keymap.go`

### Adding a New API Method
1. Add types in `internal/linear/types.go`
2. Add query in `queries.go` or mutation in `mutations.go`
3. Add message type in `internal/app/messages.go`
4. Add command in `internal/app/app.go`

### Adding Keybindings
1. Add binding to `KeyMap` struct in `keymap.go`
2. Add to `DefaultKeyMap()` function
3. Handle in appropriate view's update function

## Dependencies

| Package | Purpose |
|---------|---------|
| `bubbletea` | TUI framework (alias as `tea`) |
| `bubbles` | TUI components (spinner, textinput) |
| `lipgloss` | Styling |
| `cobra` | CLI commands |
| `viper` | Configuration |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LAZYLINER_API_KEY` | Linear API key (required) |
| `LINEAR_API_KEY` | Fallback API key |

Config file: `~/.config/lazyliner/config.yaml`
