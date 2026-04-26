package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

func newTestAdapter(t *testing.T) db.Adapter {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "kanji-comment-repo-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.Close()

	adapter, err := db.NewSQLiteAdapter(tmpFile.Name())
	if err != nil {
		t.Fatalf("new sqlite adapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })

	if err := db.RunMigrations(context.Background(), adapter.Raw()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return adapter
}

func seedTask(t *testing.T, ctx context.Context, q *sqlc.Queries) domain.Task {
	t.Helper()
	providerID := "p1"
	workspaceID := "w1"
	boardID := "b1"
	columnID := "c1"
	taskID := "t1"

	if err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        providerID,
		Type:      "local",
		Name:      "Test Provider",
		CreatedAt: "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if err := q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         workspaceID,
		ProviderID: providerID,
		Name:       "Test Workspace",
	}); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Name:        "Test Board",
		ViewDefault: "kanban",
	}); err != nil {
		t.Fatalf("create board: %v", err)
	}
	if err := q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       columnID,
		BoardID:  boardID,
		Name:     "To Do",
		Color:    "#6B7280",
		Position: 1,
	}); err != nil {
		t.Fatalf("create column: %v", err)
	}
	if err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:            taskID,
		ProviderID:    providerID,
		WorkspaceID:   workspaceID,
		BoardID:       nullString(&boardID),
		ColumnID:      nullString(&columnID),
		Title:         "Test Task",
		DescriptionMd: "",
		Priority:      0,
		LabelsJSON:    "[]",
		Position:      1,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}
	return domain.Task{ID: taskID, ProviderID: providerID}
}

func TestCommentRepository_Create(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()

	task := seedTask(t, ctx, adapter.Queries())

	repo := NewCommentRepository(adapter)
	comment := domain.Comment{
		ID:         "cm1",
		TaskID:     task.ID,
		ProviderID: task.ProviderID,
		BodyMD:     "hello world",
		Author:     strPtr("alice"),
		CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := repo.Create(ctx, comment); err != nil {
		t.Fatalf("create comment: %v", err)
	}

	comments, err := repo.ListByTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	got := comments[0]
	if got.ID != comment.ID {
		t.Errorf("ID = %q, want %q", got.ID, comment.ID)
	}
	if got.BodyMD != comment.BodyMD {
		t.Errorf("BodyMD = %q, want %q", got.BodyMD, comment.BodyMD)
	}
	if got.TaskID != comment.TaskID {
		t.Errorf("TaskID = %q, want %q", got.TaskID, comment.TaskID)
	}
	if got.Author == nil || *got.Author != "alice" {
		t.Errorf("Author = %v, want alice", got.Author)
	}
}

func strPtr(s string) *string {
	return &s
}
