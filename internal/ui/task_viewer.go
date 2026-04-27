package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/domain"
)

// scroll helpers

func scrollUp(scroll int) int {
	if scroll > 0 {
		return scroll - 1
	}
	return scroll
}

func scrollDown(scroll, maxScroll int) int {
	if scroll < maxScroll {
		return scroll + 1
	}
	return scroll
}

func pageUp(scroll, viewportLines int) int {
	if scroll > 0 {
		step := max(1, viewportLines/2)
		scroll -= step
		if scroll < 0 {
			scroll = 0
		}
	}
	return scroll
}

func pageDown(scroll, maxScroll, viewportLines int) int {
	if scroll < maxScroll {
		step := max(1, viewportLines/2)
		scroll += step
		if scroll > maxScroll {
			scroll = maxScroll
		}
	}
	return scroll
}

// task viewer lifecycle

func (m *Model) openTaskViewer() tea.Cmd {
	task, ok := m.currentTask()
	if !ok {
		return nil
	}
	return m.openTaskViewerByID(task.ID)
}

func (m *Model) openTaskViewerByID(taskID string) tea.Cmd {
	if strings.TrimSpace(taskID) == "" {
		return nil
	}
	m.overlayState.openTaskView(taskID)
	m.comments = nil
	return m.loadCommentsCmd(taskID)
}

func (m *Model) closeTaskViewer() {
	m.overlayState.closeTaskView()
}

func (m *Model) setTaskViewerReturn(taskID string) {
	m.overlayState.setTaskViewerReturn(strings.TrimSpace(taskID))
}

func (m *Model) clearTaskViewerReturn() {
	m.overlayState.clearTaskViewerReturn()
}

func (m Model) viewerTask() (domain.Task, bool) {
	if strings.TrimSpace(m.viewTaskID) != "" {
		for _, task := range m.tasks {
			if task.ID == m.viewTaskID {
				return task, true
			}
		}
	}
	return m.currentTask()
}

func (m Model) updateTaskViewer(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tasksLoadedMsg:
		return m.handleTasksLoaded(msg, false, false)
	case commentsLoadedMsg:
		return m.handleCommentsLoaded(msg)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Confirm), key.Matches(msg, m.keys.OpenDetails):
			m.closeTaskViewer()
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			m.closeTaskViewer()
			return m, nil
		case key.Matches(msg, m.keys.EditTitle):
			task, ok := m.viewerTask()
			if !ok {
				return m, nil
			}
			m.setTaskViewerReturn(task.ID)
			m.closeTaskViewer()
			return m.enterEditTaskForm(task)
		case key.Matches(msg, m.keys.AddComment):
			task, ok := m.viewerTask()
			if !ok {
				return m, nil
			}
			m.setTaskViewerReturn(task.ID)
			m.closeTaskViewer()
			return m, m.startAddComment()
		case key.Matches(msg, m.keys.Up):
			m.viewDescScroll = scrollUp(m.viewDescScroll)
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.viewDescScroll = scrollDown(m.viewDescScroll, m.taskViewerMaxDescScroll())
			return m, nil
		}

		switch msg.String() {
		case "pgup":
			m.viewDescScroll = pageUp(m.viewDescScroll, m.taskViewerDescViewportLines())
			return m, nil
		case "pgdown":
			m.viewDescScroll = pageDown(m.viewDescScroll, m.taskViewerMaxDescScroll(), m.taskViewerDescViewportLines())
			return m, nil
		case "home":
			m.viewDescScroll = 0
			return m, nil
		case "end":
			m.viewDescScroll = m.taskViewerMaxDescScroll()
			return m, nil
		}
	}
	return m, nil
}
