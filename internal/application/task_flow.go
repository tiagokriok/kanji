package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

type TaskFlow struct {
	repo domain.TaskRepository
}

func NewTaskFlow(repo domain.TaskRepository) *TaskFlow {
	return &TaskFlow{repo: repo}
}

func (f *TaskFlow) ListTasks(ctx context.Context, filters ListTaskFilters) ([]domain.Task, error) {
	if strings.TrimSpace(filters.WorkspaceID) == "" {
		return nil, errors.New("workspace id is required")
	}
	return f.repo.List(ctx, domain.TaskFilter{
		WorkspaceID: filters.WorkspaceID,
		BoardID:     filters.BoardID,
		TitleQuery:  strings.TrimSpace(filters.TitleQuery),
		ColumnID:    strings.TrimSpace(filters.ColumnID),
		Status:      strings.TrimSpace(filters.Status),
		DueSoonBy:   filters.DueSoonBy(time.Now().UTC()),
	})
}

type AdjacentMoveResult struct {
	TaskID   string
	ColumnID string
	Status   string
	Message  string
}

func (f *TaskFlow) MoveTaskAdjacent(ctx context.Context, taskID string, columns []domain.Column, currentColumnID *string, delta int) (AdjacentMoveResult, error) {
	if len(columns) == 0 {
		return AdjacentMoveResult{}, errors.New("no columns available")
	}

	current := 0
	if currentColumnID != nil {
		for i, col := range columns {
			if col.ID == *currentColumnID {
				current = i
				break
			}
		}
	}

	next := (current + delta) % len(columns)
	if next < 0 {
		next += len(columns)
	}

	col := columns[next]
	columnID := col.ID
	status := strings.ToLower(col.Name)

	err := f.MoveTask(ctx, taskID, &columnID, &status, float64(time.Now().UTC().UnixNano()))
	if err != nil {
		return AdjacentMoveResult{}, err
	}

	return AdjacentMoveResult{
		TaskID:   taskID,
		ColumnID: columnID,
		Status:   status,
		Message:  fmt.Sprintf("moved to %s", col.Name),
	}, nil
}

func (f *TaskFlow) MoveTask(ctx context.Context, taskID string, columnID, status *string, position float64) error {
	if strings.TrimSpace(taskID) == "" {
		return errors.New("task id is required")
	}
	if position == 0 {
		position = float64(time.Now().UTC().UnixNano())
	}
	return f.repo.Move(ctx, domain.MoveTaskInput{
		TaskID:    taskID,
		ColumnID:  trimStringPointer(columnID),
		Status:    trimStringPointer(status),
		Position:  position,
		UpdatedAt: time.Now().UTC(),
	})
}
