package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openFilterPanel() {
	m.overlayState.openFilters()
}

func (m *Model) closeFilterPanel() {
	m.overlayState.closeFilters()
}

func (m Model) statusFilterLabel() string {
	if m.filterIndex >= 0 && m.filterIndex < len(m.columns) {
		return m.columns[m.filterIndex].Name
	}
	if strings.TrimSpace(m.columnFilter) != "" {
		return m.columnName(m.columnFilter)
	}
	return "All"
}

func (m Model) dueFilterLabel() string {
	switch m.dueFilter {
	case dueFilterSoon:
		return "Due in 7d"
	case dueFilterOverdue:
		return "Overdue"
	case dueFilterNoDate:
		return "No due date"
	default:
		return "Any"
	}
}

func (m Model) priorityFilterLabel() string {
	switch m.priorityFilter {
	case 0:
		return "Critical (0)"
	case 1:
		return "Urgent (1)"
	case 2:
		return "High (2)"
	case 3:
		return "Medium (3)"
	case 4:
		return "Low (4)"
	case 5:
		return "None (5)"
	default:
		return "All"
	}
}

func (m Model) sortModeLabel() string {
	switch m.sortMode {
	case sortByDueDate:
		return "Due date"
	case sortByTitle:
		return "Title"
	case sortByUpdated:
		return "Updated"
	case sortByCreated:
		return "Created"
	default:
		return "Priority"
	}
}

func (m *Model) setStatusFilterByIndex(index int) {
	m.filterIndex = index
	if index < 0 || index >= len(m.columns) {
		m.filterIndex = -1
		m.columnFilter = ""
		return
	}
	m.columnFilter = m.columns[index].ID
}

func (m *Model) cycleDueFilter() {
	next := int(m.dueFilter) + 1
	if next > int(dueFilterNoDate) {
		next = int(dueFilterAny)
	}
	m.dueFilter = dueFilterMode(next)
}

func (m *Model) cycleSortMode() {
	next := int(m.sortMode) + 1
	if next > int(sortByCreated) {
		next = int(sortByPriority)
	}
	m.sortMode = taskSortMode(next)
}

func (m *Model) cycleColumnFilter() {
	if len(m.columns) == 0 {
		m.setStatusFilterByIndex(-1)
		return
	}
	m.setStatusFilterByIndex(m.filterIndex + 1)
}

func (m *Model) adjustFilterSelection(delta int) (bool, error) {
	changed := false
	switch m.filterFocus {
	case 0: // workspace
		if len(m.workspaces) == 0 {
			return false, nil
		}
		current := workspaceIndexByID(m.workspaces, m.workspaceID)
		if current < 0 {
			current = 0
		}
		next := (current + delta + len(m.workspaces)) % len(m.workspaces)
		if m.workspaces[next].ID != m.workspaceID {
			if err := m.switchWorkspace(m.workspaces[next].ID); err != nil {
				return false, err
			}
			changed = true
		}
	case 1: // board
		if len(m.boards) == 0 {
			return false, nil
		}
		current := boardIndexByID(m.boards, m.boardID)
		if current < 0 {
			current = 0
		}
		next := (current + delta + len(m.boards)) % len(m.boards)
		if m.boards[next].ID != m.boardID {
			if err := m.switchBoard(m.boards[next].ID); err != nil {
				return false, err
			}
			changed = true
		}
	case 2: // status
		total := len(m.columns) + 1 // + All
		if total <= 0 {
			return false, nil
		}
		current := m.filterIndex + 1
		next := (current + delta + total) % total
		m.setStatusFilterByIndex(next - 1)
		changed = true
	case 3: // due
		total := int(dueFilterNoDate) + 1
		next := (int(m.dueFilter) + delta + total) % total
		m.dueFilter = dueFilterMode(next)
		changed = true
	case 4: // priority
		total := 7 // all + 0..5
		current := m.priorityFilter + 1
		next := (current + delta + total) % total
		m.priorityFilter = next - 1
		changed = true
	case 5: // sort
		total := int(sortByCreated) + 1
		next := (int(m.sortMode) + delta + total) % total
		m.sortMode = taskSortMode(next)
		changed = true
	}
	return changed, nil
}

func (m Model) updateFilterPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		return m.handleTasksLoaded(msg, false, true)
	case commentsLoadedMsg:
		return m.handleCommentsLoaded(msg)
	case opResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusLine = msg.err.Error()
			return m, nil
		}
		m.statusLine = ""
		return m, m.loadTasksCmd()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.ShowFilters):
			m.closeFilterPanel()
			return m, nil
		case key.Matches(msg, m.keys.Up):
			m.filterFocus--
			if m.filterFocus < 0 {
				m.filterFocus = 5
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.filterFocus++
			if m.filterFocus > 5 {
				m.filterFocus = 0
			}
			return m, nil
		case key.Matches(msg, m.keys.Left):
			if changed, err := m.adjustFilterSelection(-1); err != nil {
				m.statusLine = err.Error()
				return m, nil
			} else if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
		case key.Matches(msg, m.keys.Right):
			if changed, err := m.adjustFilterSelection(1); err != nil {
				m.statusLine = err.Error()
				return m, nil
			} else if changed {
				return m, m.loadTasksCmd()
			}
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			m.closeFilterPanel()
			return m, nil
		}
	}
	return m, nil
}

func (m Model) renderFilterPanel(base string) string {
	_ = base
	panelWidth := m.width * 2 / 3
	if panelWidth < 56 {
		panelWidth = 56
	}
	if panelWidth > m.width-2 {
		panelWidth = max(20, m.width-2)
	}
	panelHeight := 12
	if panelHeight > m.height-2 {
		panelHeight = max(8, m.height-2)
	}

	contentWidth := boxContentWidth(panelWidth, 1, true)
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("151")).Render("Filter & Sort"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("↑/↓ select row | ←/→ change value | Enter/Esc close"),
		"",
	}

	row := func(index int, label, value string) string {
		left := lipgloss.NewStyle().Width(12).Foreground(lipgloss.Color("246")).Render(label)
		right := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
		line := left + " " + right
		if m.filterFocus == index {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Render(line)
		}
		return line
	}

	lines = append(lines,
		row(0, "Workspace", m.workspaceName),
		row(1, "Board", m.boardName),
		row(2, "Status", m.statusFilterLabel()),
		row(3, "Due", m.dueFilterLabel()),
		row(4, "Priority", m.priorityFilterLabel()),
		row(5, "Sort", m.sortModeLabel()),
	)

	panel := lipgloss.NewStyle().
		Width(contentWidth).
		Height(boxContentHeight(panelHeight, true)).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("151")).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}
