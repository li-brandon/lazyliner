# Copilot Instructions — Lazyliner

Purpose: Enable AI agents to be immediately productive in this Go + Bubble Tea TUI for Linear.

## Big Picture

- **CLI + TUI split**: Cobra commands drive entrypoints in [cmd/lazyliner/main.go](cmd/lazyliner/main.go); the interactive TUI runs `app.Model` in [internal/app/app.go](internal/app/app.go).
- **Model-driven views**: `app.Model` manages `ViewList`, `ViewDetail`, `ViewCreate`, `ViewHelp` and tabs (`My Issues`, `All`, `Active`, `Backlog`). View code lives under [internal/ui/views/\*\*](internal/ui/views).
- **Async data flow**: Use Bubble Tea commands that return typed messages (see [internal/app/messages.go](internal/app/messages.go)). Example: `loadIssues()` returns `IssuesLoadedMsg`; `Update()` pattern consumes messages and updates `Model` state.
- **Linear client**: GraphQL queries/mutations live in [internal/linear/queries.go](internal/linear/queries.go) and [internal/linear/mutations.go](internal/linear/mutations.go); shared HTTP + `execute()` in [internal/linear/client.go](internal/linear/client.go). Labels are a connection; `convertIssues()` flattens `Labels.Nodes` into `Issue.Labels`.
- **Config + integrations**: App config via Viper in [internal/config/config.go](internal/config/config.go). Git/clipboard/browser helpers in [internal/git/git.go](internal/git/git.go).

## Developer Workflows

- **Build/run**: `make build`, `make run`, `make install`. Hot-reload dev: `make dev` (installs `air` if missing).
- **Quality**: `make fmt`, `make lint` (installs `golangci-lint`), `make tidy`.
- **Tests**: `make test`, coverage: `make test-coverage`. Example single test: `go test -v ./internal/linear -run TestGetIssues`.
- **Release builds**: `make release` produces multi-platform binaries in `bin/`.

## Project Conventions

- **Imports order**: stdlib, internal packages, then external; alias Bubble Tea as `tea` (see AGENTS.md).
- **Enums**: Use `iota` for `View`/`Tab` in `app.go`.
- **Errors**: Wrap with context (`fmt.Errorf("...: %w", err)`) and check nullable fields (e.g., `issue.State`).
- **Keybindings**: Centralized in [internal/app/keymap.go](internal/app/keymap.go); vim-style + arrows; `?` toggles help.
- **Styling**: Lipgloss styles/colors in [internal/ui/theme/styles.go](internal/ui/theme/styles.go) and [internal/ui/theme/colors.go](internal/ui/theme/colors.go).

## External Integration Points

- **Linear API**: `Client.execute(ctx, query, vars, &result)` handles auth header and errors; base URL is `https://api.linear.app/graphql`.
- **Config**: Read from `~/.config/lazyliner/config.yaml` or env. Env prefix `LAZYLINER_` with fallback `LINEAR_API_KEY`. See [internal/config/config.go](internal/config/config.go).
- **Clipboard/Browser**: macOS uses `pbcopy`/`open`. Linux uses `xclip`/`xdg-open`. Windows uses `clip`/shell. See [internal/git/git.go](internal/git/git.go).

## Adding Features (follow these patterns)

- **New View**:
  1. Create model under [internal/ui/views/](internal/ui/views). 2) Add enum in `View` (in [internal/app/app.go](internal/app/app.go)). 3) Handle in `Update()` and `View()`. 4) Add navigation in `keymap.go`.
- **New API Method**:
  1. Define types in [internal/linear/types.go](internal/linear/types.go). 2) Add query/mutation in `queries.go`/`mutations.go`. 3) Create message in [internal/app/messages.go](internal/app/messages.go). 4) Wire a `tea.Cmd` in `app.Model`.
- **New Keybinding**: Add to `KeyMap` + `DefaultKeyMap()` in [internal/app/keymap.go](internal/app/keymap.go), handle in the relevant view update.
- **Issue updates**: Prefer helpers like `UpdateIssueState`, `UpdateIssueAssignee`, `UpdateIssuePriority`, `UpdateIssueLabels` in [internal/linear/mutations.go](internal/linear/mutations.go).

## Concrete Examples

- **Message + Command**:
  - Message: `type IssuesLoadedMsg struct { Issues []linear.Issue; Err error }` (see [internal/app/messages.go](internal/app/messages.go)).
  - Command: `func (m Model) loadIssues() tea.Cmd { ... return IssuesLoadedMsg{Issues: issues, Err: err} }` (see [internal/app/app.go](internal/app/app.go)).
- **GraphQL Query**: `GetTeams()` uses `execute()` with `Teams { nodes { id name key } }` (see [internal/linear/client.go](internal/linear/client.go)).
- **Config YAML**:
  ```yaml
  linear:
    api_key: lin_api_xxxxx
  defaults:
    view: my-issues
  git:
    branch_prefix: feature
    branch_format: "{prefix}/{id}-{title}"
  ```

## Gotchas

- **Labels connection**: Always access labels via `Labels.Nodes` in raw results; rely on `convertIssues()` to flatten.
- **Search mode**: List view has a search input; ensure `textinput.Blink` when focusing (see `updateListView()` in `app.go`).
- **Window sizing**: Views use `SetSize(width, height-4)` to reserve header/footer space; preserve this when adding new views.

Feedback: If any section is unclear or missing, point to the file and workflow you’re using, and we’ll refine this doc.
