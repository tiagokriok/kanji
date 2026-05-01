package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/application"
)

// ── AssembleCreateTaskInput ──

func TestAssembleCreateTaskInput_Basic(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "My Task", "--description", "desc", "--priority", "high", "--due-date", "2025-12-25", "--labels", "a,b"}))

	input, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.NoError(t, err)
	assert.Equal(t, "ws-1", input.WorkspaceID)
	assert.Equal(t, "board-1", *input.BoardID)
	assert.Equal(t, "col-1", *input.ColumnID)
	assert.Equal(t, "My Task", input.Title)
	assert.Equal(t, "desc", input.DescriptionMD)
	assert.Equal(t, 2, input.Priority)
	assert.NotNil(t, input.DueAt)
	assert.Equal(t, []string{"a", "b"}, input.Labels)
}

func TestAssembleCreateTaskInput_DefaultPriority(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "Task"}))

	input, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.NoError(t, err)
	assert.Equal(t, 3, input.Priority)
}

func TestAssembleCreateTaskInput_NoDueDate(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "Task"}))

	input, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.NoError(t, err)
	assert.Nil(t, input.DueAt)
}

func TestAssembleCreateTaskInput_LabelsNormalization(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "Task", "--labels", " A , a , B "}))

	input, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, input.Labels)
}

func TestAssembleCreateTaskInput_InvalidPriority(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("priority", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "Task", "--priority", "invalid"}))

	_, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "priority")
}

func TestAssembleCreateTaskInput_InvalidDueDate(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("due-date", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "Task", "--due-date", "bad"}))

	_, err := AssembleCreateTaskInput(cmd, "ws-1", "board-1", "col-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "due date")
}

// ── AssembleUpdateTaskInput ──

func TestAssembleUpdateTaskInput_OnlyChangedFields(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--title", "New Title"}))

	input, err := AssembleUpdateTaskInput(cmd)
	require.NoError(t, err)
	assert.Equal(t, "New Title", *input.Title)
	assert.Nil(t, input.DescriptionMD)
	assert.Nil(t, input.Priority)
	assert.Nil(t, input.DueAt)
	assert.Nil(t, input.Labels)
}

func TestAssembleUpdateTaskInput_ClearDescription(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--clear-description"}))

	input, err := AssembleUpdateTaskInput(cmd)
	require.NoError(t, err)
	assert.Equal(t, "", *input.DescriptionMD)
}

func TestAssembleUpdateTaskInput_ClearDueDate(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--clear-due-date"}))

	input, err := AssembleUpdateTaskInput(cmd)
	require.NoError(t, err)
	assert.Nil(t, input.DueAt)
	assert.True(t, input.ClearDueAt)
}

func TestAssembleUpdateTaskInput_ClearLabels(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--clear-labels"}))

	input, err := AssembleUpdateTaskInput(cmd)
	require.NoError(t, err)
	require.NotNil(t, input.Labels)
	assert.Empty(t, *input.Labels)
}

func TestAssembleUpdateTaskInput_MutualExclusion(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--description", "x", "--clear-description"}))

	_, err := AssembleUpdateTaskInput(cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// ── ResolveTaskID ──

func TestResolveTaskID_ByID(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Find Me",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("task-id", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--task-id", task.ID}))

	id, err := ResolveTaskID(cmd, rt2, "")
	require.NoError(t, err)
	assert.Equal(t, task.ID, id)
}

func TestResolveTaskID_ByTitle(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Find Me",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("task", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--task", "Find Me"}))

	id, err := ResolveTaskID(cmd, rt2, setup.Workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, id)
}

func TestResolveTaskID_Missing(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("task-id", "", "")
	cmd.Flags().String("task", "", "")

	_, err = ResolveTaskID(cmd, rt2, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task-id or task")
}

// ── ResolveMoveDestination ──

func TestResolveTaskID_Ambiguous(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	_, err = rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Duplicate",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	_, err = rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Duplicate",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("task", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--task", "Duplicate"}))

	_, err = ResolveTaskID(cmd, rt2, setup.Workspace.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestResolveMoveDestination_ByID(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("to-column-id", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--to-column-id", setup.Columns[1].ID}))

	colID, status, err := ResolveMoveDestination(cmd, rt2, setup.Board.ID)
	require.NoError(t, err)
	assert.Equal(t, setup.Columns[1].ID, colID)
	assert.Equal(t, strings.ToLower(setup.Columns[1].Name), status)
}

func TestResolveMoveDestination_ByName(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("to-column", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--to-column", setup.Columns[1].Name}))

	colID, status, err := ResolveMoveDestination(cmd, rt2, setup.Board.ID)
	require.NoError(t, err)
	assert.Equal(t, setup.Columns[1].ID, colID)
	assert.Equal(t, strings.ToLower(setup.Columns[1].Name), status)
}

func TestResolveMoveDestination_NotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("to-column-id", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--to-column-id", "invalid"}))

	_, _, err = ResolveMoveDestination(cmd, rt2, setup.Board.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ── task create ──

func TestTaskCreate_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().String("column-id", "", "")
	cmd.Flags().String("column", "", "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--workspace-id", setup.Workspace.ID,
		"--board-id", setup.Board.ID,
		"--title", "New Task",
		"--priority", "high",
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task created")
	assert.Contains(t, output, "New Task")
}

func TestTaskCreate_MissingTitle(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("title", "", "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--workspace-id", setup.Workspace.ID,
		"--board-id", setup.Board.ID,
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestTaskCreate_MissingBoard(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("board-id", "", "")
	cmd.Flags().String("board", "", "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--workspace-id", setup.Workspace.ID,
		"--title", "Task",
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestTaskCreate_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--workspace-id", setup.Workspace.ID,
		"--board-id", setup.Board.ID,
		"--title", "JSON Task",
		"--json",
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "task")
	assert.Contains(t, output, "JSON Task")
}

// ── task update ──

func TestTaskUpdate_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Old Title",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
		"--title", "New Title",
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task updated")
}

func TestTaskUpdate_MissingTask(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("title", "", "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--title", "x",
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskUpdate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task-id or task")
}

func TestTaskUpdate_NoPatchFields(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Title",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskUpdate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
}

func TestTaskUpdate_ClearDueDateSetsNull(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	due := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Dated",
		Status:      &setup.Columns[0].Name,
		DueAt:       &due,
	})
	require.NoError(t, err)
	require.NotNil(t, task.DueAt)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("description-file", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("due-date", "", "")
	cmd.Flags().StringSlice("labels", nil, "")
	cmd.Flags().Bool("clear-description", false, "")
	cmd.Flags().Bool("clear-due-date", false, "")
	cmd.Flags().Bool("clear-labels", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
		"--clear-due-date",
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskUpdate(cmd, ns)
	require.NoError(t, err)

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	updated, err := rt2.TaskService.GetTask(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.DueAt)
}

// ── task move ──

func TestTaskMove_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Movable",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("to-column-id", setup.Columns[1].ID, "")
	cmd.Flags().String("to-column", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
		"--to-column-id", setup.Columns[1].ID,
		"--board-id", setup.Board.ID,
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskMove(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task moved")
}

func TestTaskMove_MissingTask(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("to-column-id", "", "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--to-column-id", "col-2",
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskMove(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task-id or task")
}

func TestTaskMove_MissingDestination(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Movable",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskMove(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to-column-id or to-column")
}

func TestTaskMove_CrossBoardBlocked(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	// Create a second board in the same workspace.
	board2, err := rt.ContextService.CreateBoardWithColumns(context.Background(), setup.Workspace.ID, "Board 2", []application.CreateBoardColumnInput{
		{Name: "Col A", Color: "#FF0000"},
		{Name: "Col B", Color: "#00FF00"},
	})
	require.NoError(t, err)
	cols2, err := rt.ContextService.ListColumns(context.Background(), board2.ID)
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Movable",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cfg2 := RuntimeConfig{DBPath: dbPath}
	rt2, err := NewRuntime(context.Background(), cfg2)
	require.NoError(t, err)
	defer rt2.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("to-column-id", "", "")
	cmd.Flags().String("to-column", cols2[1].Name, "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("board-id", board2.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
		"--to-column", cols2[1].Name,
		"--workspace-id", setup.Workspace.ID,
		"--board-id", board2.ID,
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskMove(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot move task to a different board")
}

// ── task delete ──

func TestTaskDelete_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Deletable",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
		"--yes",
	}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskDelete(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Task deleted")
}

func TestTaskDelete_MissingYes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	task, err := rt.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &setup.Columns[0].ID,
		Title:       "Deletable",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--task-id", task.ID,
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskDelete(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "yes")
}

func TestTaskDelete_MissingTask(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{
		"--db-path", dbPath,
		"--yes",
	}))

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runTaskDelete(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task-id or task")
}
