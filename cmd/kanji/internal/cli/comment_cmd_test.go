package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/application"
)

func TestCommentList_Basic(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)

	_, err = rt.CommentService.AddComment(context.Background(), application.AddCommentInput{
		TaskID:     task.ID,
		ProviderID: setup.Provider.ID,
		BodyMD:     "Test comment",
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Test comment")
}

func TestCommentList_MissingScope(t *testing.T) {
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
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentList(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task")
}

func TestCommentList_JSON(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "comments")
}

func TestCommentCreate_Basic(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--body", "Great work!"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "comment created")
	assert.Contains(t, output, task.ID)
	assert.Contains(t, output, "Great work!")
}

func TestCommentCreate_WithAuthor(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--body", "Nice!", "--author", "Alice"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "comment created")
	assert.Contains(t, output, "Alice")
}

func TestCommentCreate_BodyFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	bodyFile := filepath.Join(dir, "comment.md")
	require.NoError(t, os.WriteFile(bodyFile, []byte("File comment body"), 0o644))

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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--body-file", bodyFile}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "comment created")
	assert.Contains(t, output, "File comment body")
}

func TestCommentCreate_MissingTaskID(t *testing.T) {
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
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--body", "oops"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task")
}

func TestCommentCreate_MissingBody(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body")
}

func TestCommentCreate_MutuallyExclusiveBody(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--body", "inline", "--body-file", "some.txt"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestCommentCreate_TaskNotFound(t *testing.T) {
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
	cmd.Flags().String("task-id", "invalid-task", "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", "invalid-task", "--body", "text"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCommentCreate_JSON(t *testing.T) {
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
		Title:       "Test Task",
		Status:      &setup.Columns[0].Name,
	})
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("task-id", task.ID, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("body-file", "", "")
	cmd.Flags().String("author", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--task-id", task.ID, "--body", "JSON comment", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runCommentCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "comment")
	assert.Contains(t, output, "JSON comment")
	assert.Contains(t, output, task.ID)
}
