# 🦥 Lazyliner

A beautiful, keyboard-driven terminal TUI for [Linear](https://linear.app) built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **Issue Browser** - List, filter, and search issues with vim-style navigation
- **Issue Detail View** - Full issue details with markdown rendering
- **Issue Creation** - Interactive form to create new issues
- **Kanban Board** - Visual board view with drag-and-drop style keyboard navigation
- **Quick Actions** - Change status, assignee, priority, and labels with keyboard shortcuts
- **Multiple Views** - My Issues, All Issues, Active, and Backlog tabs
- **Linear-inspired Design** - Beautiful color scheme matching Linear's aesthetic

## Installation

### Homebrew (macOS)

```bash
brew install brandonli/tap/lazyliner
```

### From Source

```bash
# Clone the repository
git clone https://github.com/brandonli/lazyliner.git
cd lazyliner

# Build
make build

# Install to GOPATH/bin
make install
```

### Using Go

```bash
go install github.com/brandonli/lazyliner/cmd/lazyliner@latest
```

## Configuration

### API Key

Lazyliner requires a Linear API key. You can get one from [Linear Settings > API](https://linear.app/settings/api).

Set it via environment variable:

```bash
export LAZYLINER_API_KEY=lin_api_xxxxx
```

Or create a config file at `~/.config/lazyliner/config.yaml`:

```yaml
linear:
  api_key: lin_api_xxxxx

defaults:
  view: my-issues  # my-issues, all, active, backlog

ui:
  vim_mode: true
  show_ids: true

git:
  branch_prefix: feature
  branch_format: "{prefix}/{id}-{title}"
```

## Usage

```bash
# Start the TUI (default)
lazyliner

# List issues in terminal
lazyliner list
lazyliner list --mine        # Show only my issues
lazyliner list -n 50         # Show 50 issues

# View a specific issue
lazyliner view ABC-123

# Create a new issue (opens TUI)
lazyliner create
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` / `Home` | Go to top |
| `G` / `End` | Go to bottom |
| `Ctrl+d` | Page down |
| `Ctrl+u` | Page up |

### Tabs

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1` | My Issues |
| `2` | All Issues |
| `3` | Active |
| `4` | Backlog |

### Actions

| Key | Action |
|-----|--------|
| `Enter` | View issue detail |
| `/` | Search/filter issues |
| `b` | Kanban board view |
| `c` | Create new issue |
| `s` | Change status |
| `a` | Change assignee |
| `p` | Change priority |
| `l` | Manage labels |
| `y` | Copy branch name |
| `o` | Open in browser |
| `r` | Refresh |
| `?` | Toggle help |
| `q` | Quit |

### In Detail View

| Key | Action |
|-----|--------|
| `Esc` | Back to list |
| `e` | Edit issue |
| `s` | Change status |
| `a` | Change assignee |
| `p` | Change priority |

### Kanban Board

| Key | Action |
|-----|--------|
| `b` | Switch to board view (from list) |
| `h` / `←` | Move to left column |
| `l` / `→` | Move to right column |
| `j` / `↓` | Move down in column |
| `k` / `↑` | Move up in column |
| `H` | Move issue to left column |
| `L` | Move issue to right column |
| `m` | Enter move mode (then h/l or 1-9) |
| `Enter` | View issue detail |
| `Esc` | Back to list view |

## Roadmap

### MVP (Current)
- [x] Issue list view with vim navigation
- [x] Issue detail view
- [x] Issue creation form
- [x] Quick actions (status, assignee, priority, labels)
- [x] Tab navigation (My Issues, All, Active, Backlog)

### Phase 2 - Polish
- [x] Search/filter functionality
- [x] Git branch detection
- [x] Clipboard support (copy branch name)
- [x] Open in browser support
- [x] Help overlay
- [x] CLI commands (list, view, create)

### Phase 3 - Kanban
- [x] Kanban board view
- [x] Move issues between columns (keyboard-based)

### Phase 4 - Repository Analyzer
- [ ] TODO/FIXME scanner
- [ ] Test coverage analysis
- [ ] Dependency analysis
- [ ] Auto-generate issues from repo structure

### Phase 5 - AI Integration
- [ ] Configurable AI provider (OpenAI, Anthropic, Ollama)
- [ ] AI-powered issue generation from prompts

## Tech Stack

- **Language**: Go (see `go.mod` for required version)
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **CLI**: [Cobra](https://github.com/spf13/cobra)
- **Config**: [Viper](https://github.com/spf13/viper)

## License

MIT
