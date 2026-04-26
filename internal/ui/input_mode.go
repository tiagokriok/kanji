package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

// startSearch enters title-search input mode, restoring the previous filter value.
func (m *Model) startSearch() tea.Cmd {
	m.inputMode = inputSearch
	m.textInput.SetValue(m.titleFilter)
	m.textInput.Placeholder = "Search title"
	m.textInput.Focus()
	m.statusLine = "Search by title"
	return textinput.Blink
}

// cancelInput resets any active non-task-form input mode and blurs widgets.
func (m *Model) cancelInput() {
	m.inputMode = inputNone
	m.textInput.Blur()
	m.textArea.Blur()
	m.statusLine = ""
}

// confirmSearch exits search mode, stores the query, and triggers a reload.
func (m *Model) confirmSearch() tea.Cmd {
	value := strings.TrimSpace(m.textInput.Value())
	m.cancelInput()
	m.titleFilter = value
	return m.loadTasksCmd()
}

// startAddComment enters comment-input mode with an empty text input.
func (m *Model) startAddComment() tea.Cmd {
	m.inputMode = inputAddComment
	m.textInput.SetValue("")
	m.textInput.Placeholder = "Comment body"
	m.textInput.Focus()
	m.statusLine = "Add comment"
	return textinput.Blink
}

// confirmAddComment validates the comment and submits it, or reopens the input if empty.
func (m *Model) confirmAddComment() tea.Cmd {
	value := strings.TrimSpace(m.textInput.Value())
	task, ok := m.currentTask()
	previousStatus := m.statusLine
	m.cancelInput()
	if !ok {
		return nil
	}
	if value == "" {
		m.inputMode = inputAddComment
		m.textInput.Focus()
		m.statusLine = "comment is required"
		return textinput.Blink
	}
	m.statusLine = previousStatus
	return m.addCommentCmd(task.ID, value)
}

// cancelAddComment exits comment mode and returns to the task viewer if requested.
func (m *Model) cancelAddComment() tea.Cmd {
	m.cancelInput()
	if m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
		taskID := m.returnTaskID
		m.clearTaskViewerReturn()
		return m.openTaskViewerByID(taskID)
	}
	return nil
}

// startEditDescription enters inline description editing for the given task.
func (m *Model) startEditDescription(task domain.Task) {
	m.inputMode = inputEditDescription
	m.textArea.SetValue(task.DescriptionMD)
	m.textArea.Focus()
	m.statusLine = "Edit description (ctrl+s to save, esc to cancel)"
}

// submitEditDescription saves the textarea content as the task description.
func (m *Model) submitEditDescription() tea.Cmd {
	task, ok := m.currentTask()
	if !ok {
		m.inputMode = inputNone
		return nil
	}
	description := m.textArea.Value()
	previousStatus := m.statusLine
	m.cancelInput()
	m.statusLine = previousStatus
	return m.updateTaskDescriptionCmd(task.ID, description)
}
