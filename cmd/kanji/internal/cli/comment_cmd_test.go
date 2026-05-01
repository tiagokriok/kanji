package cli

import (
	"context"
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
