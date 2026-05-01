package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
	"github.com/tiagokriok/kanji/internal/infrastructure/repositories"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

func TestBoardDeleteService_BoardDeleteImpact_EmptyBoard(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	_, workspaceID := seedWorkspace(t, ctx, q)
	boardID, _ := seedBoardWithColumns(t, ctx, q, workspaceID)

	s := store.New(adapter)
	svc := NewBoardDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	impact, err := svc.BoardDeleteImpact(ctx, workspaceID, boardID)
	require.NoError(t, err)

	assert.Equal(t, boardID, impact.BoardID)
	assert.Equal(t, 2, impact.Columns)
	assert.Equal(t, 0, impact.Tasks)
	assert.Equal(t, 0, impact.Comments)
}

func TestBoardDeleteService_BoardDeleteImpact_WithData(t *testing.T) {
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
	svc := NewBoardDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	impact, err := svc.BoardDeleteImpact(ctx, workspaceID, boardID)
	require.NoError(t, err)

	assert.Equal(t, boardID, impact.BoardID)
	assert.Equal(t, 2, impact.Columns)
	assert.Equal(t, 2, impact.Tasks)
	assert.Equal(t, 2, impact.Comments)
}

func TestBoardDeleteService_BoardDeleteImpact_EmptyBoardID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewBoardDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	_, err := svc.BoardDeleteImpact(ctx, "ws-test", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board id is required")
}

func TestBoardDeleteService_Delete(t *testing.T) {
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
	svc := NewBoardDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	err := svc.DeleteBoard(ctx, boardID)
	require.NoError(t, err)

	boards, err := q.ListBoards(ctx, workspaceID)
	require.NoError(t, err)
	assert.Empty(t, boards)

	columns, err := q.ListColumns(ctx, boardID)
	require.NoError(t, err)
	assert.Empty(t, columns)

	tasks, err := q.ListTasks(ctx, sqlc.ListTasksParams{WorkspaceID: workspaceID})
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestBoardDeleteService_Delete_EmptyBoardID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewBoardDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		repositories.NewCommentRepository(s),
	)

	err := svc.DeleteBoard(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board id is required")
}
