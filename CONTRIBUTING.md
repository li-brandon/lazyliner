# Contributing to Lazyliner

Thank you for your interest in contributing to Lazyliner! This guide will help you get started with development and understand our contribution workflow.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25+** - Check your version with `go version`
- **Git** - For version control
- **Make** - For build commands (optional but recommended)
- **Linear API Key** - Get one from [Linear Settings > API](https://linear.app/settings/api)

## Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/lazyliner.git
cd lazyliner

# Add upstream remote
git remote add upstream https://github.com/brandonli/lazyliner.git
```

### 2. Set Up Your Environment

```bash
# Install dependencies
go mod download

# Set your Linear API key
export LAZYLINER_API_KEY=lin_api_xxxxx

# Or create a config file at ~/.config/lazyliner/config.yaml
mkdir -p ~/.config/lazyliner
echo "linear:
  api_key: lin_api_xxxxx" > ~/.config/lazyliner/config.yaml
```

### 3. Verify Your Setup

```bash
# Build the project
make build

# Run tests
make test

# Run the TUI
make run
```

## Development Workflow

### Creating a Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. **Study the architecture** - Read [CLAUDE.md](./CLAUDE.md) to understand the codebase structure
2. **Follow code conventions** - See the [Code Style](#code-style) section below
3. **Write tests** - Add tests for new functionality
4. **Test your changes** - Run `make test` and `make lint`
5. **Update documentation** - Update README.md, CLAUDE.md, or CONTRIBUTING.md if needed

### Development Commands

```bash
make build          # Build binary to bin/lazyliner
make run            # Run in development mode
make test           # Run all tests
make lint           # Run golangci-lint (auto-installs if missing)
make fmt            # Format code with gofmt -s
make tidy           # go mod tidy
make dev            # Hot-reload dev mode (uses air)

# Run specific tests
go test -v ./internal/linear -run TestGetIssues

# Test with coverage
make test-coverage  # Generates coverage.html
```

### Committing Your Changes

```bash
# Stage your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add keyboard shortcut for filtering by label"

# Push to your fork
git push origin feature/your-feature-name
```

#### Commit Message Format

We follow conventional commit format:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

Examples:
```
feat: add kanban board view
fix: resolve issue with label filtering
docs: update keybinding documentation
refactor: simplify issue list rendering
```

## Code Style

### Import Order

Always organize imports in three groups:

```go
import (
    "context"                                          // stdlib first
    "fmt"

    "github.com/brandonli/lazyliner/internal/config"   // internal second

    tea "github.com/charmbracelet/bubbletea"           // external third
    "github.com/charmbracelet/lipgloss"
)
```

### Naming Conventions

- **Packages**: lowercase, single word (`linear`, `theme`, `config`)
- **Exported types/functions**: PascalCase (`GetIssues`, `IssueCreateInput`)
- **Private functions**: camelCase (`buildFilter`, `convertIssues`)
- **JSON tags**: snake_case (`json:"branch_name"`)
- **Config tags**: snake_case (`mapstructure:"api_key"`)

### Error Handling

Always wrap errors with context and check nullable fields:

```go
if err != nil {
    return nil, fmt.Errorf("failed to fetch issues: %w", err)
}

// Check nullable fields from API responses
status := "Unknown"
if issue.State != nil {
    status = issue.State.Name
}
```

### Code Organization

- Follow the existing package structure (see [CLAUDE.md](./CLAUDE.md#package-structure))
- Keep functions focused and single-purpose
- Use descriptive variable names
- Add comments for exported functions and complex logic
- Avoid magic numbers - use named constants

## Testing Requirements

### Writing Tests

- **Unit tests**: Test individual functions and methods
- **Test coverage**: Aim for at least 70% coverage for new code
- **Table-driven tests**: Use for testing multiple scenarios

Example test structure:

```go
func TestGetIssues(t *testing.T) {
    tests := []struct {
        name    string
        filter  IssueFilter
        wantErr bool
    }{
        {
            name:    "fetch all issues",
            filter:  IssueFilter{},
            wantErr: false,
        },
        {
            name:    "fetch with filter",
            filter:  IssueFilter{State: "In Progress"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/linear -run TestGetIssues

# Run tests with race detector
go test -race ./...
```

### Before Submitting

Ensure all checks pass:

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Build project
make build
```

## Pull Request Process

### 1. Prepare Your PR

- Ensure your branch is up to date with main:
  ```bash
  git checkout main
  git pull upstream main
  git checkout feature/your-feature-name
  git rebase main
  ```

- Run all quality checks:
  ```bash
  make fmt lint test build
  ```

### 2. Submit Your PR

1. Push your branch to your fork
2. Open a Pull Request on GitHub
3. Fill out the PR template with:
   - **Description**: What does this PR do?
   - **Motivation**: Why is this change needed?
   - **Testing**: How did you test this?
   - **Screenshots**: For UI changes (if applicable)

### 3. PR Review Process

- A maintainer will review your PR
- Address any feedback or requested changes
- Keep your PR updated with main if needed
- Once approved, a maintainer will merge your PR

### PR Checklist

- [ ] Code follows the style guide
- [ ] Tests added for new functionality
- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow conventional format
- [ ] PR description is clear and complete

## Project Architecture

For detailed architecture information, see [CLAUDE.md](./CLAUDE.md). Key points:

- **Bubble Tea**: TUI framework with Model-Update-View pattern
- **Async operations**: Commands return typed messages
- **Package structure**: Clear separation between app, ui, linear API, and config
- **Message-driven**: UI updates through message passing

## Common Tasks

### Adding a New View

1. Create model in `internal/ui/views/your-view/`
2. Add enum value to `View` in `internal/app/app.go`
3. Handle in `Update()` and `View()` switch statements
4. Add navigation keybinding in `keymap.go`

See [CLAUDE.md - Adding Features](./CLAUDE.md#adding-features) for details.

### Adding a New API Method

1. Add types in `internal/linear/types.go`
2. Add query in `queries.go` or mutation in `mutations.go`
3. Add message type in `internal/app/messages.go`
4. Wire a `tea.Cmd` in `internal/app/app.go`

### Adding a New Keybinding

1. Add to `KeyMap` struct in `keymap.go`
2. Add to `DefaultKeyMap()` function
3. Handle in the appropriate view's update function

## Getting Help

- **Documentation**: Read [README.md](./README.md) and [CLAUDE.md](./CLAUDE.md)
- **Issues**: Check [existing issues](https://github.com/brandonli/lazyliner/issues) or create a new one
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Code examples**: Look at existing code for patterns and conventions

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## License

By contributing to Lazyliner, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Lazyliner! 🦥
