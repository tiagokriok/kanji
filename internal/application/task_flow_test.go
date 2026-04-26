package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
)

type fakeTaskRepo struct {
	tasks   []domain.Task
	columns []domain.Column
	boards  []domain.Board

	listErr error
	moveErr error

	lastListFilter domain.TaskFilter
	lastMoveInput  domain.MoveTaskInput
}

func (r *fakeTaskRepo) Create(ctx context.Context, task domain.Task) error { return nil }
func (r *fakeTaskRepo) Update(ctx context.Context, taskID string, patch domain.TaskPatch) error {
	return nil
}
func (r *fakeTaskRepo) GetByID(ctx context.Context, taskID string) (domain.Task, error) {
	return domain.Task{}, nil
}
func (r *fakeTaskRepo) List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	r.lastListFilter = filter
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.tasks, nil
}
func (r *fakeTaskRepo) Move(ctx context.Context, input domain.MoveTaskInput) error {
	r.lastMoveInput = input
	if r.moveErr != nil {
		return r.moveErr
	}
	return nil
}
func (r *fakeTaskRepo) Delete(ctx context.Context, id string) error { return nil }
func (r *fakeTaskRepo) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	return r.columns, nil
}
func (r *fakeTaskRepo) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	return r.boards, nil
}

func TestTaskFlow_ListTasks_ValidatesWorkspaceID(t *testing.T) {
	repo := &fakeTaskRepo{}
	flow := NewTaskFlow(repo)

	_, err := flow.ListTasks(context.Background(), ListTaskFilters{
		WorkspaceID: "",
	})
	if err == nil {
		t.Fatal("expected error for empty workspace id, got nil")
	}
	if err.Error() != "workspace id is required" {
		t.Fatalf("expected 'workspace id is required', got %q", err.Error())
	}
}

func TestTaskFlow_ListTasks_DelegatesToRepo(t *testing.T) {
	repo := &fakeTaskRepo{
		tasks: []domain.Task{
			{ID: "t1", Title: "Task One"},
		},
	}
	flow := NewTaskFlow(repo)

	tasks, err := flow.ListTasks(context.Background(), ListTaskFilters{
		WorkspaceID: "ws-1",
		BoardID:     "board-1",
		TitleQuery:  "  search  ",
		ColumnID:    "col-1",
		Status:      "  doing  ",
		DueSoonDays: 7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "t1" {
		t.Fatalf("expected [t1], got %v", tasks)
	}

	f := repo.lastListFilter
	if f.WorkspaceID != "ws-1" {
		t.Errorf("WorkspaceID = %q, want %q", f.WorkspaceID, "ws-1")
	}
	if f.BoardID != "board-1" {
		t.Errorf("BoardID = %q, want %q", f.BoardID, "board-1")
	}
	if f.TitleQuery != "search" {
		t.Errorf("TitleQuery = %q, want %q", f.TitleQuery, "search")
	}
	if f.ColumnID != "col-1" {
		t.Errorf("ColumnID = %q, want %q", f.ColumnID, "col-1")
	}
	if f.Status != "doing" {
		t.Errorf("Status = %q, want %q", f.Status, "doing")
	}
	if f.DueSoonBy == nil {
		t.Fatal("expected DueSoonBy to be set")
	}
}

func TestTaskFlow_ListTasks_ReturnsRepoError(t *testing.T) {
	repo := &fakeTaskRepo{listErr: errors.New("db down")}
	flow := NewTaskFlow(repo)

	_, err := flow.ListTasks(context.Background(), ListTaskFilters{WorkspaceID: "ws-1"})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected 'db down' error, got %v", err)
	}
}

func TestTaskFlow_MoveTask_ValidatesTaskID(t *testing.T) {
	repo := &fakeTaskRepo{}
	flow := NewTaskFlow(repo)

	err := flow.MoveTask(context.Background(), "", nil, nil, 1.0)
	if err == nil {
		t.Fatal("expected error for empty task id, got nil")
	}
	if err.Error() != "task id is required" {
		t.Fatalf("expected 'task id is required', got %q", err.Error())
	}
}

func TestTaskFlow_MoveTask_SetsDefaultPosition(t *testing.T) {
	repo := &fakeTaskRepo{}
	flow := NewTaskFlow(repo)

	before := time.Now().UTC().Add(-time.Second)
	err := flow.MoveTask(context.Background(), "task-1", strPtr("col-1"), strPtr("done"), 0)
	after := time.Now().UTC().Add(time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := repo.lastMoveInput
	if m.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", m.TaskID, "task-1")
	}
	if m.ColumnID == nil || *m.ColumnID != "col-1" {
		t.Errorf("ColumnID = %v, want col-1", m.ColumnID)
	}
	if m.Status == nil || *m.Status != "done" {
		t.Errorf("Status = %v, want done", m.Status)
	}
	if m.Position == 0 {
		t.Error("expected non-zero default position")
	}
	if m.UpdatedAt.Before(before) || m.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt = %v, expected between %v and %v", m.UpdatedAt, before, after)
	}
}

func TestTaskFlow_MoveTask_UsesProvidedPosition(t *testing.T) {
	repo := &fakeTaskRepo{}
	flow := NewTaskFlow(repo)

	err := flow.MoveTask(context.Background(), "task-1", nil, nil, 42.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastMoveInput.Position != 42.0 {
		t.Errorf("Position = %v, want 42.0", repo.lastMoveInput.Position)
	}
}

func TestTaskFlow_MoveTask_ReturnsRepoError(t *testing.T) {
	repo := &fakeTaskRepo{moveErr: errors.New("conflict")}
	flow := NewTaskFlow(repo)

	err := flow.MoveTask(context.Background(), "task-1", nil, nil, 1.0)
	if err == nil || err.Error() != "conflict" {
		t.Fatalf("expected 'conflict' error, got %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}
