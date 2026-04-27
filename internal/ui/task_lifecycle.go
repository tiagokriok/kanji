package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTasksLoaded processes a tasksLoadedMsg, updating tasks with active filters and sort.
// When restoreKanban is true, it attempts to restore pending kanban selection before
// falling back to ensureSelection. When refreshDetails is true, it loads comments for
// the current task if details are visible, or clears cached comments otherwise.
func (m Model) handleTasksLoaded(msg tasksLoadedMsg, restoreKanban, refreshDetails bool) (Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.statusLine = msg.err.Error()
		return m, nil
	}
	m.tasks = m.applyActiveFilters(msg.tasks)
	m.sortTasks(m.tasks)
	if restoreKanban {
		if !m.restorePendingKanbanSelection() {
			m.ensureSelection()
		}
	} else {
		m.ensureSelection()
	}
	if refreshDetails {
		return m, m.refreshDetails()
	}
	return m, nil
}

// loadCommentsIfVisible returns a command to load comments for the current task
// when showDetails is true, otherwise returns nil without mutating state.
func (m Model) loadCommentsIfVisible() tea.Cmd {
	if m.showDetails {
		if task, ok := m.currentTask(); ok {
			return m.loadCommentsCmd(task.ID)
		}
	}
	return nil
}

// refreshDetails returns a command to load comments when showDetails is true and a
// current task exists. Otherwise it clears cached comments and returns nil.
func (m *Model) refreshDetails() tea.Cmd {
	if m.showDetails {
		if task, ok := m.currentTask(); ok {
			return m.loadCommentsCmd(task.ID)
		}
	}
	m.comments = nil
	return nil
}

// handleCommentsLoaded processes a commentsLoadedMsg, updating the cached comments.
func (m Model) handleCommentsLoaded(msg commentsLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.statusLine = msg.err.Error()
		return m, nil
	}
	m.comments = msg.comments
	return m, nil
}

// handleOpResult processes an opResultMsg, updating kanban selection and handling
// task viewer return transitions. On success it triggers a task reload.
func (m Model) handleOpResult(msg opResultMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.statusLine = msg.err.Error()
		m.clearTaskViewerReturn()
		return m, nil
	}
	m.statusLine = ""
	if m.viewMode == viewKanban && strings.TrimSpace(msg.taskID) != "" && strings.TrimSpace(msg.columnID) != "" {
		m.pendingKanbanTaskID = msg.taskID
		m.pendingKanbanColumnID = msg.columnID
		m.setActiveColumnByID(msg.columnID)
	}
	if m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
		taskID := m.returnTaskID
		m.clearTaskViewerReturn()
		commentsCmd := m.openTaskViewerByID(taskID)
		return m, tea.Batch(m.loadTasksCmd(), commentsCmd)
	}
	return m, m.loadTasksCmd()
}
