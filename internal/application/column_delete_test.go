package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/infrastructure/repositories"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

func TestColumnDeleteService_ColumnTaskCount(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID := seedWorkspace(t, ctx, q)
	boardID, colIDs := seedBoardWithColumns(t, ctx, q, workspaceID)
	seedTask(t, ctx, q, providerID, workspaceID, boardID, colIDs[0])

	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	count, err := svc.ColumnTaskCount(ctx, workspaceID, colIDs[0])
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = svc.ColumnTaskCount(ctx, workspaceID, colIDs[1])
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestColumnDeleteService_ColumnTaskCount_EmptyColumnID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	_, err := svc.ColumnTaskCount(ctx, "ws-test", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "column id is required")
}

func TestColumnDeleteService_ReassignTasks(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	providerID, workspaceID := seedWorkspace(t, ctx, q)
	boardID, colIDs := seedBoardWithColumns(t, ctx, q, workspaceID)
	taskID := seedTask(t, ctx, q, providerID, workspaceID, boardID, colIDs[0])

	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	err := svc.ReassignTasks(ctx, workspaceID, colIDs[0], colIDs[1], "doing")
	require.NoError(t, err)

	task, err := q.GetTask(ctx, taskID)
	require.NoError(t, err)
	require.True(t, task.ColumnID.Valid)
	assert.Equal(t, colIDs[1], task.ColumnID.String)
	require.True(t, task.Status.Valid)
	assert.Equal(t, "doing", task.Status.String)
}

func TestColumnDeleteService_ReassignTasks_EmptyIDs(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	err := svc.ReassignTasks(ctx, "ws-test", "", "col-2", "todo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "from column id is required")

	err = svc.ReassignTasks(ctx, "ws-test", "col-1", "", "todo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to column id is required")

	err = svc.ReassignTasks(ctx, "ws-test", "col-1", "col-2", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "destination column status cannot be empty")
}

func TestColumnDeleteService_DeleteColumn(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	q := adapter.Queries()
	_, workspaceID := seedWorkspace(t, ctx, q)
	boardID, colIDs := seedBoardWithColumns(t, ctx, q, workspaceID)

	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	err := svc.DeleteColumn(ctx, colIDs[0])
	require.NoError(t, err)

	columns, err := q.ListColumns(ctx, boardID)
	require.NoError(t, err)
	assert.Len(t, columns, 1)
	assert.Equal(t, colIDs[1], columns[0].ID)
}

func TestColumnDeleteService_DeleteColumn_EmptyColumnID(t *testing.T) {
	adapter := newTestDB(t)
	ctx := context.Background()
	s := store.New(adapter)
	svc := NewColumnDeleteService(
		repositories.NewSetupRepository(s),
		repositories.NewTaskRepository(s),
		NewTaskFlow(repositories.NewTaskRepository(s)),
	)

	err := svc.DeleteColumn(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "column id is required")
}
