package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

type TaskRepository struct {
	db      db.Adapter
	queries *sqlc.Queries
}

func NewTaskRepository(adapter db.Adapter) *TaskRepository {
	return &TaskRepository{
		db:      adapter,
		queries: adapter.Queries(),
	}
}

func (r *TaskRepository) Create(ctx context.Context, task domain.Task) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for create task: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
	err = qtx.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:            task.ID,
		ProviderID:    task.ProviderID,
		WorkspaceID:   task.WorkspaceID,
		BoardID:       nullString(task.BoardID),
		ColumnID:      nullString(task.ColumnID),
		RemoteID:      nullString(task.RemoteID),
		Title:         task.Title,
		DescriptionMd: task.DescriptionMD,
		Status:        nullString(task.Status),
		Priority:      int64(task.Priority),
		DueAt:         nullableTimeToString(task.DueAt),
		EstimateMinutes: func() sql.NullInt64 {
			if task.EstimateMinutes == nil {
				return sql.NullInt64{}
			}
			return sql.NullInt64{Int64: int64(*task.EstimateMinutes), Valid: true}
		}(),
		Assignee:   nullString(task.Assignee),
		LabelsJSON: marshalLabels(task.Labels),
		Position:   task.Position,
		CreatedAt:  task.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  task.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Update(ctx context.Context, taskID string, patch domain.TaskPatch) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for update task: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
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
	if err := qtx.UpdateTask(ctx, arg); err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID string) (domain.Task, error) {
	item, err := r.queries.GetTask(ctx, taskID)
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

	items, err := r.queries.ListTasks(ctx, arg)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for move task: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)
	err = qtx.MoveTask(ctx, sqlc.MoveTaskParams{
		ColumnID:  nullString(input.ColumnID),
		Status:    nullString(input.Status),
		Position:  input.Position,
		UpdatedAt: input.UpdatedAt.UTC().Format(time.RFC3339),
		ID:        input.TaskID,
	})
	if err != nil {
		return fmt.Errorf("move task: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit move task: %w", err)
	}
	return nil
}

func (r *TaskRepository) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	items, err := r.queries.ListColumns(ctx, boardID)
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
	items, err := r.queries.ListBoards(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Board, 0, len(items))
	for _, item := range items {
		result = append(result, fromSQLBoard(item))
	}
	return result, nil
}
