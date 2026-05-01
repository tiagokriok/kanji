package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

// ColumnDeleteService handles column deletion and task reassignment.
type ColumnDeleteService struct {
	setupRepo domain.SetupRepository
	taskRepo  domain.TaskRepository
	taskFlow  *TaskFlow
}

// NewColumnDeleteService creates a new ColumnDeleteService.
func NewColumnDeleteService(setup domain.SetupRepository, task domain.TaskRepository, flow *TaskFlow) *ColumnDeleteService {
	return &ColumnDeleteService{setupRepo: setup, taskRepo: task, taskFlow: flow}
}

// ColumnTaskCount returns the number of tasks in a column.
func (s *ColumnDeleteService) ColumnTaskCount(ctx context.Context, workspaceID, columnID string) (int, error) {
	columnID = strings.TrimSpace(columnID)
	if columnID == "" {
		return 0, fmt.Errorf("column id is required")
	}

	tasks, err := s.taskRepo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, ColumnID: columnID})
	if err != nil {
		return 0, err
	}
	return len(tasks), nil
}

// ReassignTasks moves all tasks from one column to another.
func (s *ColumnDeleteService) ReassignTasks(ctx context.Context, workspaceID, fromColumnID, toColumnID, toStatus string) error {
	fromColumnID = strings.TrimSpace(fromColumnID)
	toColumnID = strings.TrimSpace(toColumnID)
	if fromColumnID == "" {
		return fmt.Errorf("from column id is required")
	}
	if toColumnID == "" {
		return fmt.Errorf("to column id is required")
	}
	if toStatus == "" {
		return errors.New("destination column status cannot be empty")
	}

	tasks, err := s.taskRepo.List(ctx, domain.TaskFilter{WorkspaceID: workspaceID, ColumnID: fromColumnID})
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := s.taskFlow.MoveTask(ctx, task.ID, &toColumnID, &toStatus, 0); err != nil {
			return err
		}
	}
	return nil
}

// DeleteColumn removes a column.
func (s *ColumnDeleteService) DeleteColumn(ctx context.Context, columnID string) error {
	columnID = strings.TrimSpace(columnID)
	if columnID == "" {
		return fmt.Errorf("column id is required")
	}

	return s.setupRepo.DeleteColumn(ctx, columnID)
}
