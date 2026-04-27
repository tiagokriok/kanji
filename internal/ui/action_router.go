package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type executeActionMsg struct {
	action string
}

type keybindEntry struct {
	ID    string
	Key   string
	Label string
}

func (m Model) keybindEntries() []keybindEntry {
	entries := []keybindEntry{
		{ID: "new_task", Key: "n", Label: "Create task"},
		{ID: "edit_task", Key: "e", Label: "Edit selected task"},
		{ID: "edit_description", Key: "E", Label: "Edit description"},
		{ID: "add_comment", Key: "c", Label: "Add comment"},
		{ID: "search", Key: "/", Label: "Search"},
		{ID: "open_filters", Key: "f", Label: "Open filter/sort panel"},
		{ID: "open_workspaces", Key: "w", Label: "Open workspace switcher"},
		{ID: "open_board_panel", Key: "b", Label: "Open board manager"},
		{ID: "prev_board", Key: "[", Label: "Previous board"},
		{ID: "next_board", Key: "]", Label: "Next board"},
		{ID: "toggle_details", Key: "d", Label: "Toggle details pane"},
		{ID: "open_move", Key: "Enter", Label: "Open task viewer"},
		{ID: "move_task", Key: "m", Label: "Move task to next status"},
		{ID: "toggle_view", Key: "Tab", Label: "Switch list/kanban"},
		{ID: "cycle_status", Key: "s", Label: "Cycle status filter"},
		{ID: "cycle_due_filter", Key: "z", Label: "Cycle due filter"},
		{ID: "cycle_sort", Key: "o", Label: "Cycle sort mode"},
		{ID: "move_up", Key: "↑", Label: "Move selection up"},
		{ID: "move_down", Key: "↓", Label: "Move selection down"},
		{ID: "move_left", Key: "←", Label: "Move selection left (kanban)"},
		{ID: "move_right", Key: "→", Label: "Move selection right (kanban)"},
		{ID: "move_task_left", Key: "Shift+←", Label: "Move card to left column (kanban)"},
		{ID: "move_task_right", Key: "Shift+→", Label: "Move card to right column (kanban)"},
		{ID: "quit", Key: "q", Label: "Quit"},
	}
	if strings.TrimSpace(m.titleFilter) != "" {
		entries = append(entries, keybindEntry{ID: "clear_search", Key: "x", Label: "Clear search"})
	}
	return entries
}

func (m Model) filteredKeybindEntries() []keybindEntry {
	query := strings.ToLower(strings.TrimSpace(m.keyFilter.Value()))
	entries := m.keybindEntries()
	if query == "" {
		return entries
	}

	filtered := make([]keybindEntry, 0, len(entries))
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Key), query) || strings.Contains(strings.ToLower(entry.Label), query) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (m Model) executeAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "quit":
		return m, tea.Quit
	case "toggle_view":
		if m.viewMode == viewList {
			m.viewMode = viewKanban
		} else {
			m.viewMode = viewList
		}
		m.ensureSelection()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "toggle_details":
		m.showDetails = !m.showDetails
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "search":
		return m, m.startSearch()
	case "open_filters":
		m.openFilterPanel()
		return m, nil
	case "open_workspaces":
		m.openContextPanel(contextWorkspace)
		return m, textinput.Blink
	case "open_board_panel":
		m.openContextPanel(contextBoard)
		return m, textinput.Blink
	case "prev_board":
		changed, err := m.switchBoardByOffset(-1)
		if err != nil {
			m.statusLine = err.Error()
			return m, nil
		}
		if changed {
			return m, m.loadTasksCmd()
		}
		return m, nil
	case "next_board":
		changed, err := m.switchBoardByOffset(1)
		if err != nil {
			m.statusLine = err.Error()
			return m, nil
		}
		if changed {
			return m, m.loadTasksCmd()
		}
		return m, nil
	case "clear_search":
		if strings.TrimSpace(m.titleFilter) == "" {
			return m, nil
		}
		m.titleFilter = ""
		m.statusLine = ""
		return m, m.loadTasksCmd()
	case "new_task":
		return m.enterCreateTaskForm()
	case "edit_task":
		task, ok := m.currentTask()
		if !ok {
			return m, nil
		}
		return m.enterEditTaskForm(task)
	case "add_comment":
		if _, ok := m.currentTask(); !ok {
			return m, nil
		}
		return m, m.startAddComment()
	case "edit_description":
		task, ok := m.currentTask()
		if !ok {
			return m, nil
		}
		return m, m.startExternalDescriptionEdit(task)
	case "cycle_status":
		m.cycleColumnFilter()
		return m, m.loadTasksCmd()
	case "cycle_due_filter":
		m.cycleDueFilter()
		return m, m.loadTasksCmd()
	case "cycle_sort":
		m.cycleSortMode()
		return m, m.loadTasksCmd()
	case "move_up":
		m.moveUp()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "move_down":
		m.moveDown()
		if m.showDetails {
			if task, ok := m.currentTask(); ok {
				return m, m.loadCommentsCmd(task.ID)
			}
		}
		return m, nil
	case "move_left":
		if m.viewMode == viewKanban {
			if m.activeColumn > 0 {
				m.activeColumn--
			}
			m.ensureKanbanRow()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
		}
		return m, nil
	case "move_right":
		if m.viewMode == viewKanban {
			if m.activeColumn < len(m.columns)-1 {
				m.activeColumn++
			}
			m.ensureKanbanRow()
			if m.showDetails {
				if task, ok := m.currentTask(); ok {
					return m, m.loadCommentsCmd(task.ID)
				}
			}
		}
		return m, nil
	case "move_task_left":
		if m.viewMode == viewKanban {
			if task, ok := m.currentTask(); ok {
				return m, m.moveToPrevColumnCmd(task)
			}
		}
		return m, nil
	case "move_task_right":
		if m.viewMode == viewKanban {
			if task, ok := m.currentTask(); ok {
				return m, m.moveToNextColumnCmd(task)
			}
		}
		return m, nil
	case "open_move":
		return m, m.openTaskViewer()
	case "move_task":
		if task, ok := m.currentTask(); ok {
			return m, m.moveToNextColumnCmd(task)
		}
		return m, nil
	default:
		return m, nil
	}
}
