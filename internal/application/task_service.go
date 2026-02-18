package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tiagokriok/lazytask/internal/domain"
)

type CreateTaskInput struct {
	ProviderID    string
	WorkspaceID   string
	BoardID       *string
	ColumnID      *string
	Title         string
	DescriptionMD string
	Status        *string
	Priority      int
	DueAt         *time.Time
	Labels        []string
}

type UpdateTaskInput struct {
	Title         *string
	DescriptionMD *string
	Status        *string
	Priority      *int
	DueAt         *time.Time
	ColumnID      *string
	Labels        *[]string
}

type TaskService struct {
	repo domain.TaskRepository
}

func NewTaskService(repo domain.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) (domain.Task, error) {
	if strings.TrimSpace(input.ProviderID) == "" {
		return domain.Task{}, errors.New("provider id is required")
	}
	if strings.TrimSpace(input.WorkspaceID) == "" {
		return domain.Task{}, errors.New("workspace id is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return domain.Task{}, errors.New("title is required")
	}

	now := time.Now().UTC()
	task := domain.Task{
		ID:            uuid.NewString(),
		ProviderID:    input.ProviderID,
		WorkspaceID:   input.WorkspaceID,
		BoardID:       input.BoardID,
		ColumnID:      input.ColumnID,
		Title:         strings.TrimSpace(input.Title),
		DescriptionMD: input.DescriptionMD,
		Status:        input.Status,
		Priority:      input.Priority,
		DueAt:         input.DueAt,
		Labels:        normalizeLabels(input.Labels),
		Position:      float64(now.UnixNano()),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID string, input UpdateTaskInput) error {
	if strings.TrimSpace(taskID) == "" {
		return errors.New("task id is required")
	}
	patch := domain.TaskPatch{
		Title:         trimStringPointer(input.Title),
		DescriptionMD: input.DescriptionMD,
		Status:        trimStringPointer(input.Status),
		Priority:      input.Priority,
		DueAt:         input.DueAt,
		ColumnID:      trimStringPointer(input.ColumnID),
		Labels:        normalizeLabelPatch(input.Labels),
	}
	return s.repo.Update(ctx, taskID, patch)
}

func (s *TaskService) MoveTask(ctx context.Context, taskID string, columnID, status *string, position float64) error {
	if strings.TrimSpace(taskID) == "" {
		return errors.New("task id is required")
	}
	if position == 0 {
		position = float64(time.Now().UTC().UnixNano())
	}
	return s.repo.Move(ctx, domain.MoveTaskInput{
		TaskID:    taskID,
		ColumnID:  trimStringPointer(columnID),
		Status:    trimStringPointer(status),
		Position:  position,
		UpdatedAt: time.Now().UTC(),
	})
}

func (s *TaskService) ListTasks(ctx context.Context, filters ListTaskFilters) ([]domain.Task, error) {
	if strings.TrimSpace(filters.WorkspaceID) == "" {
		return nil, errors.New("workspace id is required")
	}
	return s.repo.List(ctx, domain.TaskFilter{
		WorkspaceID: filters.WorkspaceID,
		BoardID:     filters.BoardID,
		TitleQuery:  strings.TrimSpace(filters.TitleQuery),
		ColumnID:    strings.TrimSpace(filters.ColumnID),
		Status:      strings.TrimSpace(filters.Status),
		DueSoonBy:   filters.DueSoonBy(time.Now().UTC()),
	})
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (domain.Task, error) {
	if strings.TrimSpace(taskID) == "" {
		return domain.Task{}, errors.New("task id is required")
	}
	return s.repo.GetByID(ctx, taskID)
}

func (s *TaskService) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	if strings.TrimSpace(boardID) == "" {
		return nil, errors.New("board id is required")
	}
	return s.repo.ListColumns(ctx, boardID)
}

func (s *TaskService) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, errors.New("workspace id is required")
	}
	return s.repo.ListBoards(ctx, workspaceID)
}

func normalizeLabels(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, v := range in {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeLabelPatch(in *[]string) *[]string {
	if in == nil {
		return nil
	}
	labels := normalizeLabels(*in)
	return &labels
}

func trimStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	return &v
}
