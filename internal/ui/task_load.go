package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
)

// loadTasksCmd returns a command that loads tasks for the current workspace and board
// using the active filter state. The result is delivered as a tasksLoadedMsg.
func (m Model) loadTasksCmd() tea.Cmd {
	filters := application.ListTaskFilters{
		WorkspaceID: m.workspaceID,
		BoardID:     m.boardID,
		TitleQuery:  m.titleFilter,
		ColumnID:    m.columnFilter,
	}
	flow := m.taskFlow
	return func() tea.Msg {
		tasks, err := flow.ListTasks(context.Background(), filters)
		return tasksLoadedMsg{tasks: tasks, err: err}
	}
}

// loadCommentsCmd returns a command that loads comments for the given task ID.
// The result is delivered as a commentsLoadedMsg.
func (m Model) loadCommentsCmd(taskID string) tea.Cmd {
	service := m.commentService
	return func() tea.Msg {
		comments, err := service.ListComments(context.Background(), taskID)
		return commentsLoadedMsg{comments: comments, err: err}
	}
}

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
