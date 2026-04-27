package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

type TaskRepository struct {
	store store.Store
}

func NewTaskRepository(s store.Store) *TaskRepository {
	return &TaskRepository{store: s}
}

func (r *TaskRepository) Create(ctx context.Context, task domain.Task) error {
	return r.store.Write(ctx, "create task", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.CreateTask(ctx, sqlc.CreateTaskParams{
			ID:              task.ID,
			ProviderID:      task.ProviderID,
			WorkspaceID:     task.WorkspaceID,
			BoardID:         nullString(task.BoardID),
			ColumnID:        nullString(task.ColumnID),
			RemoteID:        nullString(task.RemoteID),
			Title:           task.Title,
			DescriptionMd:   task.DescriptionMD,
			Status:          nullString(task.Status),
			Priority:        int64(task.Priority),
			DueAt:           nullableTimeToString(task.DueAt),
			EstimateMinutes: nullInt(task.EstimateMinutes),
			Assignee:        nullString(task.Assignee),
			LabelsJSON:      marshalLabels(task.Labels),
			Position:        task.Position,
			CreatedAt:       task.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:       task.UpdatedAt.UTC().Format(time.RFC3339),
		})
	})
}

func (r *TaskRepository) Update(ctx context.Context, taskID string, patch domain.TaskPatch) error {
	return r.store.Write(ctx, "update task", func(tx store.Tx) error {
		qtx := tx.Queries()
		arg := sqlc.UpdateTaskParams{
			Title:         nullString(patch.Title),
			DescriptionMd: nullString(patch.DescriptionMD),
			Status:        nullString(patch.Status),
			Priority:      nullInt(patch.Priority),
			DueAt:         nullableTimeToString(patch.DueAt),
			ColumnID:      nullString(patch.ColumnID),
			UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
			ID:            taskID,
		}
		if patch.Labels != nil {
			arg.LabelsJSON = sql.NullString{String: marshalLabels(*patch.Labels), Valid: true}
		}
		return qtx.UpdateTask(ctx, arg)
	})
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID string) (domain.Task, error) {
	item, err := r.store.Queries().GetTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return fromSQLTask(item), nil
}

func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	arg := sqlc.ListTasksParams{
		WorkspaceID: filter.WorkspaceID,
		BoardID:     filter.BoardID,
		TitleQuery:  filter.TitleQuery,
		ColumnID:    filter.ColumnID,
		Status:      filter.Status,
	}
	if filter.DueSoonBy != nil {
		arg.DueSoonActive = 1
		arg.DueSoonBefore = filter.DueSoonBy.UTC().Format(time.RFC3339)
	}

	items, err := r.store.Queries().ListTasks(ctx, arg)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Task, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLTask(item))
	}
	return result, nil
}

func (r *TaskRepository) Move(ctx context.Context, input domain.MoveTaskInput) error {
	return r.store.Write(ctx, "move task", func(tx store.Tx) error {
		qtx := tx.Queries()
		return qtx.MoveTask(ctx, sqlc.MoveTaskParams{
			ColumnID:  nullString(input.ColumnID),
			Status:    nullString(input.Status),
			Position:  input.Position,
			UpdatedAt: input.UpdatedAt.UTC().Format(time.RFC3339),
			ID:        input.TaskID,
		})
	})
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	return r.store.Write(ctx, "delete task", func(tx store.Tx) error {
		return tx.Queries().DeleteTask(ctx, id)
	})
}

func (r *TaskRepository) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	items, err := r.store.Queries().ListColumns(ctx, boardID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Column, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLColumn(item))
	}
	return result, nil
}

func (r *TaskRepository) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	items, err := r.store.Queries().ListBoards(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Board, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLBoard(item))
	}
	return result, nil
}
