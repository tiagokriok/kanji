package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// dispatchOverlayUpdate routes messages to the active overlay updater.
// It returns (model, cmd, true) when an overlay consumed the message.
func (m Model) dispatchOverlayUpdate(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch m.activeOverlay() {
	case overlayTaskView:
		model, cmd := m.updateTaskViewer(msg)
		return model, cmd, true
	case overlayKeybinds:
		model, cmd := m.updateKeybindPanel(msg)
		return model, cmd, true
	case overlayFilters:
		model, cmd := m.updateFilterPanel(msg)
		return model, cmd, true
	case overlayContexts:
		model, cmd := m.updateContextPanel(msg)
		return model, cmd, true
	case overlayInput:
		model, cmd := m.updateInputMode(msg)
		return model, cmd, true
	default:
		return m, nil, false
	}
}

// dispatchGlobalMessage handles global messages outside overlays.
// It returns (model, cmd, true) when the message was handled.
func (m Model) dispatchGlobalMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textArea.SetWidth(max(20, msg.Width/2-6))
		m.keyFilter.Width = max(24, msg.Width/3)
		return m, nil, true
	case executeActionMsg:
		model, cmd := m.executeAction(msg.action)
		return model, cmd, true
	case tasksLoadedMsg:
		model, cmd := m.handleTasksLoaded(msg, true, true)
		return model, cmd, true
	case commentsLoadedMsg:
		model, cmd := m.handleCommentsLoaded(msg)
		return model, cmd, true
	case descriptionEditedMsg:
		model, cmd := m.handleExternalDescriptionEdited(msg)
		return model, cmd, true
	case opResultMsg:
		model, cmd := m.handleOpResult(msg)
		return model, cmd, true
	}
	return m, nil, false
}

// handleDeleteConfirmKey handles delete confirmation keys.
// It returns (model, cmd, true) when confirmingDelete was active.
func (m Model) handleDeleteConfirmKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.confirmingDelete {
		return m, nil, false
	}
	if msg.String() == "y" {
		if task, ok := m.currentTask(); ok {
			m.confirmingDelete = false
			m.statusLine = ""
			return m, m.deleteTaskCmd(task.ID), true
		}
	}
	m.confirmingDelete = false
	m.statusLine = ""
	return m, nil, true
}
