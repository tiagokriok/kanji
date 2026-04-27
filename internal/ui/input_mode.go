package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
	if !ok {
		m.inputMode = inputNone
		m.textInput.Blur()
		return nil
	}
	m.cancelInput()
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

// handleDescriptionEditedMsg processes a descriptionEditedMsg while in input mode.
// It returns handled=true when the message was consumed (task form mode).
func (m Model) handleDescriptionEditedMsg(msg descriptionEditedMsg) (tea.Model, tea.Cmd, bool) {
	if m.inputMode != inputTaskForm || m.taskForm == nil {
		return m, nil, false
	}
	if msg.err != nil {
		m.statusLine = fmt.Sprintf("editor error: %v", msg.err)
		return m, nil, true
	}
	m.taskForm.descriptionFull = msg.content
	m.taskForm.description.SetValue(summarizeDescription(msg.content))
	m.statusLine = ""
	return m, nil, true
}

// cancelInputMode handles the cancel key for any active input mode.
func (m Model) cancelInputMode() (tea.Model, tea.Cmd) {
	switch m.inputMode {
	case inputTaskForm:
		m.closeTaskForm()
		if m.returnTaskView && strings.TrimSpace(m.returnTaskID) != "" {
			taskID := m.returnTaskID
			m.clearTaskViewerReturn()
			return m, m.openTaskViewerByID(taskID)
		}
		return m, nil
	case inputAddComment:
		return m, m.cancelAddComment()
	default:
		m.cancelInput()
		return m, nil
	}
}

// saveEditDescription saves the current textarea content as the task description.
func (m Model) saveEditDescription() (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		m.inputMode = inputNone
		return m, nil
	}
	description := m.textArea.Value()
	m.inputMode = inputNone
	m.textArea.Blur()
	return m, m.updateTaskDescriptionCmd(task.ID, description)
}

// confirmInputMode handles the confirm key for any active non-edit-description input mode.
// It returns handled=true when the confirm action was consumed.
func (m Model) confirmInputMode() (tea.Model, tea.Cmd, bool) {
	switch m.inputMode {
	case inputSearch:
		return m, m.confirmSearch(), true
	case inputAddComment:
		return m, m.confirmAddComment(), true
	}
	return m, nil, false
}

// updateInputModeWidgets routes non-consumed messages to the active input widget.
// updateTaskForm routes messages to the task form key handler or widget updater.
func (m Model) updateTaskForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if model, cmd, handled := m.handleTaskFormKey(keyMsg); handled {
			return model, cmd
		}
	}
	return m.updateTaskFormWidgets(msg)
}

// updateTaskFormWidgets routes non-consumed messages to the active task form field.
func (m Model) updateTaskFormWidgets(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.taskForm == nil {
		return m, nil
	}
	if field := m.taskForm.currentInputField(); field != nil {
		prevDescription := m.taskForm.description.Value()
		var cmd tea.Cmd
		*field, cmd = field.Update(msg)
		if m.taskForm.focus == taskFieldDescription && m.taskForm.description.Value() != prevDescription {
			m.taskForm.descriptionFull = m.taskForm.description.Value()
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) updateInputMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case descriptionEditedMsg:
		if model, cmd, handled := m.handleDescriptionEditedMsg(msg); handled {
			return model, cmd
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			return m.cancelInputMode()
		case m.inputMode == inputTaskForm:
			return m.updateTaskForm(msg)
		case msg.String() == "ctrl+s" && m.inputMode == inputEditDescription:
			return m.saveEditDescription()
		case key.Matches(msg, m.keys.Confirm) && m.inputMode != inputEditDescription:
			if model, cmd, handled := m.confirmInputMode(); handled {
				return model, cmd
			}
		}
	}

	if m.inputMode == inputTaskForm {
		return m.updateTaskForm(msg)
	}
	return m.updateInputModeWidgets(msg)
}

func (m Model) updateInputModeWidgets(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.inputMode {
	case inputEditDescription:
		var cmd tea.Cmd
		m.textArea, cmd = m.textArea.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}
