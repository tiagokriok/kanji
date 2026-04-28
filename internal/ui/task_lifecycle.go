package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
