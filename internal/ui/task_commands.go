package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

func (m Model) createTaskWithDetailsCmd(title, description string, priority int, dueAt *time.Time, boardID, columnID, status *string) tea.Cmd {
	service := m.taskService
	providerID := m.providerID
	workspaceID := m.workspaceID

	return func() tea.Msg {
		_, err := service.CreateTask(context.Background(), application.CreateTaskInput{
			ProviderID:    providerID,
			WorkspaceID:   workspaceID,
			BoardID:       boardID,
			ColumnID:      columnID,
			Status:        status,
			Title:         title,
			DescriptionMD: description,
			Priority:      priority,
			DueAt:         dueAt,
			Labels:        []string{},
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task created"}
	}
}

func (m Model) updateTaskWithDetailsCmd(taskID string, title, description *string, priority *int, dueAt *time.Time, columnID, status *string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.UpdateTask(context.Background(), taskID, application.UpdateTaskInput{
			Title:         title,
			DescriptionMD: description,
			Priority:      priority,
			DueAt:         dueAt,
			ColumnID:      columnID,
			Status:        status,
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task updated"}
	}
}

func (m Model) updateTaskDescriptionCmd(taskID, description string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.UpdateTask(context.Background(), taskID, application.UpdateTaskInput{DescriptionMD: &description})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "description updated"}
	}
}

func (m Model) addCommentCmd(taskID, body string) tea.Cmd {
	service := m.commentService
	providerID := m.providerID
	return func() tea.Msg {
		_, err := service.AddComment(context.Background(), application.AddCommentInput{
			TaskID:     taskID,
			ProviderID: providerID,
			BodyMD:     body,
		})
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "comment added"}
	}
}

func (m Model) moveToNextColumnCmd(task domain.Task) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	flow := m.taskFlow
	return func() tea.Msg {
		result, err := flow.MoveTaskAdjacent(context.Background(), task.ID, m.columns, task.ColumnID, 1)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: result.Message, taskID: result.TaskID, columnID: result.ColumnID}
	}
}

func (m Model) moveToPrevColumnCmd(task domain.Task) tea.Cmd {
	if len(m.columns) == 0 {
		return nil
	}
	flow := m.taskFlow
	return func() tea.Msg {
		result, err := flow.MoveTaskAdjacent(context.Background(), task.ID, m.columns, task.ColumnID, -1)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: result.Message, taskID: result.TaskID, columnID: result.ColumnID}
	}
}

func (m Model) deleteTaskCmd(id string) tea.Cmd {
	service := m.taskService
	return func() tea.Msg {
		err := service.DeleteTask(context.Background(), id)
		if err != nil {
			return opResultMsg{err: err}
		}
		return opResultMsg{status: "task deleted"}
	}
}
