package application

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/repositories"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

func newTestDB(t *testing.T) *db.SQLiteAdapter {
	t.Helper()
	tmpFile, err := db.NewSQLiteAdapter(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { tmpFile.Close() })

	if err := db.RunMigrations(context.Background(), tmpFile.Raw()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return tmpFile
}

func seedWorkspace(t *testing.T, ctx context.Context, q *sqlc.Queries) (string, string) {
	t.Helper()
	providerID := "p-test"
	err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        providerID,
		Type:      "local",
		Name:      "Test Provider",
		CreatedAt: "2024-01-01T00:00:00Z",
	})
	require.NoError(t, err)

	workspaceID := "ws-test"
	err = q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID:         workspaceID,
		ProviderID: providerID,
		Name:       "Test Workspace",
	})
	require.NoError(t, err)

	return providerID, workspaceID
}

func seedBoardWithColumns(t *testing.T, ctx context.Context, q *sqlc.Queries, workspaceID string) (string, []string) {
	t.Helper()
	boardID := "board-test"
	err := q.CreateBoard(ctx, sqlc.CreateBoardParams{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Name:        "Test Board",
		ViewDefault: "list",
	})
	require.NoError(t, err)

	col1 := "col-1"
	err = q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       col1,
		BoardID:  boardID,
		Name:     "To Do",
		Color:    "#FF0000",
		Position: 1,
	})
	require.NoError(t, err)

	col2 := "col-2"
	err = q.CreateColumn(ctx, sqlc.CreateColumnParams{
		ID:       col2,
		BoardID:  boardID,
		Name:     "Done",
		Color:    "#00FF00",
		Position: 2,
	})
	require.NoError(t, err)

	return boardID, []string{col1, col2}
}

func seedTask(t *testing.T, ctx context.Context, q *sqlc.Queries, providerID, workspaceID, boardID, columnID string) string {
	t.Helper()
	taskID := "task-" + columnID
	err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:            taskID,
		ProviderID:    providerID,
		WorkspaceID:   workspaceID,
		BoardID:       sql.NullString{String: boardID, Valid: true},
		ColumnID:      sql.NullString{String: columnID, Valid: true},
		Title:         "Test Task",
		DescriptionMd: "",
		Priority:      0,
		Position:      1,
		CreatedAt:     "2024-01-01T00:00:00Z",
		UpdatedAt:     "2024-01-01T00:00:00Z",
	})
	require.NoError(t, err)
	return taskID
}

func seedComment(t *testing.T, ctx context.Context, q *sqlc.Queries, taskID, providerID string) {
	t.Helper()
	err := q.CreateComment(ctx, sqlc.CreateCommentParams{
		ID:         "comment-" + taskID,
		TaskID:     taskID,
		ProviderID: providerID,
		BodyMd:     "Test comment",
		CreatedAt:  "2024-01-01T00:00:00Z",
	})
	require.NoError(t, err)
}

func TestWorkspaceDeleteService_Impact_EmptyWorkspace(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	_, workspaceID := seedWorkspace(t, ctx, q)

	s := store.New(adapter)
	svc := NewWorkspaceDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	impact, err := svc.Impact(ctx, workspaceID)
	require.NoError(t, err)

	assert.Equal(t, workspaceID, impact.WorkspaceID)
	assert.Equal(t, 0, impact.Boards)
	assert.Equal(t, 0, impact.Columns)
	assert.Equal(t, 0, impact.Tasks)
	assert.Equal(t, 0, impact.Comments)
}

func TestWorkspaceDeleteService_Impact_WithData(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID := seedWorkspace(t, ctx, q)
	boardID, colIDs := seedBoardWithColumns(t, ctx, q, workspaceID)

	for _, colID := range colIDs {
		taskID := seedTask(t, ctx, q, providerID, workspaceID, boardID, colID)
		seedComment(t, ctx, q, taskID, providerID)
	}

	s := store.New(adapter)
	svc := NewWorkspaceDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	impact, err := svc.Impact(ctx, workspaceID)
	require.NoError(t, err)

	assert.Equal(t, workspaceID, impact.WorkspaceID)
	assert.Equal(t, 1, impact.Boards)
	assert.Equal(t, 2, impact.Columns)
	assert.Equal(t, 2, impact.Tasks)
	assert.Equal(t, 2, impact.Comments)
}

func TestWorkspaceDeleteService_Impact_EmptyWorkspaceID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewWorkspaceDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	_, err := svc.Impact(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace id is required")
}

func TestWorkspaceDeleteService_Delete(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID := seedWorkspace(t, ctx, q)
	boardID, colIDs := seedBoardWithColumns(t, ctx, q, workspaceID)

	for _, colID := range colIDs {
		taskID := seedTask(t, ctx, q, providerID, workspaceID, boardID, colID)
		seedComment(t, ctx, q, taskID, providerID)
	}

	s := store.New(adapter)
	svc := NewWorkspaceDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	err := svc.Delete(ctx, workspaceID)
	require.NoError(t, err)

	workspaces, err := q.ListWorkspaces(ctx)
	require.NoError(t, err)
	assert.Empty(t, workspaces)

	boards, err := q.ListBoards(ctx, workspaceID)
	require.NoError(t, err)
	assert.Empty(t, boards)

	columns, err := q.ListColumns(ctx, boardID)
	require.NoError(t, err)
	assert.Empty(t, columns)

	tasks, err := q.ListTasks(ctx, sqlc.ListTasksParams{WorkspaceID: workspaceID})
	require.NoError(t, err)
	assert.Empty(t, tasks)

	var commentCount int
	err = adapter.Raw().QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)
	require.NoError(t, err)
	assert.Equal(t, 0, commentCount)
}

func TestWorkspaceDeleteService_Delete_EmptyWorkspaceID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewWorkspaceDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	err := svc.Delete(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace id is required")
}
