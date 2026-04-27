package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type executeActionMsg struct {
	action string
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
