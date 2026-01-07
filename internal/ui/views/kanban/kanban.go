package kanban

import (
	"fmt"
	"sort"

	"github.com/brandonli/lazyliner/internal/linear"
	"github.com/brandonli/lazyliner/internal/ui/theme"
	"github.com/brandonli/lazyliner/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	State  linear.WorkflowState
	Issues []linear.Issue
	Cursor int
}

type Model struct {
	columns      []Column
	activeColumn int
	width        int
	height       int
	columnWidth  int
	moveMode     bool
}

func New(issues []linear.Issue, states []linear.WorkflowState, width, height int) Model {
	sortedStates := sortStatesByType(states)

	columns := make([]Column, len(sortedStates))
	for i, state := range sortedStates {
		columns[i] = Column{
			State:  state,
			Issues: filterIssuesByState(issues, state.ID),
			Cursor: 0,
		}
	}

	colWidth := calculateColumnWidth(width, len(columns))

	return Model{
		columns:      columns,
		activeColumn: 0,
		width:        width,
		height:       height,
		columnWidth:  colWidth,
		moveMode:     false,
	}
}

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.columnWidth = calculateColumnWidth(width, len(m.columns))
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.moveMode {
			return m.updateMoveMode(msg)
		}
		return m.updateNormalMode(msg)
	}
	return m, nil
}

func (m Model) updateNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		if m.activeColumn > 0 {
			m.activeColumn--
			m.clampCursor()
		}
	case "l", "right":
		if m.activeColumn < len(m.columns)-1 {
			m.activeColumn++
			m.clampCursor()
		}
	case "j", "down":
		col := &m.columns[m.activeColumn]
		if col.Cursor < len(col.Issues)-1 {
			col.Cursor++
		}
	case "k", "up":
		col := &m.columns[m.activeColumn]
		if col.Cursor > 0 {
			col.Cursor--
		}
	case "g", "home":
		m.columns[m.activeColumn].Cursor = 0
	case "G", "end":
		col := &m.columns[m.activeColumn]
		if len(col.Issues) > 0 {
			col.Cursor = len(col.Issues) - 1
		}
	case "m":
		if m.SelectedIssue() != nil {
			m.moveMode = true
		}
	case "H":
		return m.moveIssueLeft()
	case "L":
		return m.moveIssueRight()
	}
	return m, nil
}

func (m Model) updateMoveMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.moveMode = false
	case "h", "left":
		m.moveMode = false
		return m.moveIssueLeft()
	case "l", "right":
		m.moveMode = false
		return m.moveIssueRight()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.String()[0] - '1')
		if idx >= 0 && idx < len(m.columns) {
			m.moveMode = false
			return m.moveIssueToColumn(idx)
		}
	}
	return m, nil
}

func (m Model) moveIssueLeft() (Model, tea.Cmd) {
	if m.activeColumn > 0 {
		return m.moveIssueToColumn(m.activeColumn - 1)
	}
	return m, nil
}

func (m Model) moveIssueRight() (Model, tea.Cmd) {
	if m.activeColumn < len(m.columns)-1 {
		return m.moveIssueToColumn(m.activeColumn + 1)
	}
	return m, nil
}

func (m Model) moveIssueToColumn(targetCol int) (Model, tea.Cmd) {
	issue := m.SelectedIssue()
	if issue == nil || targetCol == m.activeColumn {
		return m, nil
	}

	targetState := m.columns[targetCol].State

	return m, func() tea.Msg {
		return MoveIssueMsg{
			IssueID: issue.ID,
			StateID: targetState.ID,
		}
	}
}

func (m *Model) clampCursor() {
	col := &m.columns[m.activeColumn]
	if col.Cursor >= len(col.Issues) {
		col.Cursor = len(col.Issues) - 1
	}
	if col.Cursor < 0 {
		col.Cursor = 0
	}
}

func (m Model) SelectedIssue() *linear.Issue {
	if m.activeColumn < 0 || m.activeColumn >= len(m.columns) {
		return nil
	}
	col := m.columns[m.activeColumn]
	if col.Cursor < 0 || col.Cursor >= len(col.Issues) {
		return nil
	}
	return &col.Issues[col.Cursor]
}

func (m Model) View() string {
	if len(m.columns) == 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			theme.TextMutedStyle.Render("No workflow states available"),
		)
	}

	var cols []string
	for i, col := range m.columns {
		isActive := i == m.activeColumn
		cols = append(cols, m.renderColumn(col, isActive))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	if m.moveMode {
		hint := theme.StatusBarStyle.
			Width(m.width).
			Render("Move mode: h/l or 1-9 to select column, ESC to cancel")
		return lipgloss.JoinVertical(lipgloss.Left, board, hint)
	}

	return board
}

func (m Model) renderColumn(col Column, isActive bool) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Text).
		Background(theme.Surface).
		Padding(0, 1).
		Width(m.columnWidth - 2).
		Align(lipgloss.Center)

	if isActive {
		headerStyle = headerStyle.
			Foreground(theme.Primary).
			Background(theme.SurfaceHover)
	}

	statusIcon := theme.StatusIcon(col.State.Type)
	header := headerStyle.Render(fmt.Sprintf("%s %s (%d)", statusIcon, col.State.Name, len(col.Issues)))

	cardHeight := 4
	maxCards := (m.height - 4) / (cardHeight + 1)
	if maxCards < 1 {
		maxCards = 1
	}

	startIdx := 0
	if col.Cursor >= maxCards {
		startIdx = col.Cursor - maxCards + 1
	}

	var cards []string
	for i := startIdx; i < len(col.Issues) && i < startIdx+maxCards; i++ {
		issue := col.Issues[i]
		isSelected := isActive && i == col.Cursor
		cards = append(cards, m.renderCard(issue, isSelected))
	}

	if len(col.Issues) == 0 {
		emptyCard := lipgloss.NewStyle().
			Width(m.columnWidth-4).
			Height(cardHeight).
			Foreground(theme.TextDim).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No issues")
		cards = append(cards, emptyCard)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, cards...)

	columnStyle := lipgloss.NewStyle().
		Width(m.columnWidth).
		Height(m.height-2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(0, 1)

	if isActive {
		columnStyle = columnStyle.BorderForeground(theme.Primary)
	}

	return columnStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", content))
}

func (m Model) renderCard(issue linear.Issue, isSelected bool) string {
	cardStyle := lipgloss.NewStyle().
		Width(m.columnWidth-6).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border)

	if isSelected {
		cardStyle = cardStyle.
			BorderForeground(theme.Primary).
			Background(theme.SurfaceHover)
	}

	idStyle := theme.IssueIDStyle
	title := util.Truncate(issue.Title, m.columnWidth-10)

	priorityIcon := theme.PriorityIcon(issue.Priority)

	assignee := ""
	if issue.Assignee != nil {
		assignee = util.Truncate(issue.Assignee.Name, 15)
	}

	line1 := lipgloss.JoinHorizontal(lipgloss.Top,
		idStyle.Render(issue.Identifier),
		"  ",
		priorityIcon,
	)

	titleStyle := lipgloss.NewStyle().Foreground(theme.Text)
	if isSelected {
		titleStyle = titleStyle.Foreground(theme.TextBright)
	}
	line2 := titleStyle.Render(title)

	var line3 string
	if assignee != "" {
		line3 = theme.TextMutedStyle.Render("👤 " + assignee)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, line1, line2)
	if line3 != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, line3)
	}

	return cardStyle.Render(content)
}

type MoveIssueMsg struct {
	IssueID string
	StateID string
}

func sortStatesByType(states []linear.WorkflowState) []linear.WorkflowState {
	sorted := make([]linear.WorkflowState, len(states))
	copy(sorted, states)

	typeOrder := map[string]int{
		"backlog":   0,
		"unstarted": 1,
		"started":   2,
		"completed": 3,
		"canceled":  4,
	}

	sort.Slice(sorted, func(i, j int) bool {
		orderI := typeOrder[sorted[i].Type]
		orderJ := typeOrder[sorted[j].Type]
		if orderI != orderJ {
			return orderI < orderJ
		}
		return sorted[i].Position < sorted[j].Position
	})

	return sorted
}

func filterIssuesByState(issues []linear.Issue, stateID string) []linear.Issue {
	var filtered []linear.Issue
	for _, issue := range issues {
		if issue.State != nil && issue.State.ID == stateID {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func calculateColumnWidth(totalWidth, numColumns int) int {
	if numColumns == 0 {
		return totalWidth
	}
	colWidth := totalWidth / numColumns
	if colWidth < 30 {
		colWidth = 30
	}
	if colWidth > 50 {
		colWidth = 50
	}
	return colWidth
}
