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
	"github.com/tiagokriok/kanji/internal/state"
)

func TestColumnList_ExplicitScope(t *testing.T) {
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
	cmd.Flags().String("board-id", setup.Board.ID, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Todo")
	assert.Contains(t, output, "Doing")
	assert.Contains(t, output, "Done")
}

func TestColumnList_MissingScope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnListWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestColumnList_JSON(t *testing.T) {
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
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "columns")
	assert.Contains(t, output, "Todo")
}

func TestColumnCreate_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "Review", "--color", "#FF5733"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "column created")
	assert.Contains(t, output, "Review")
	assert.Contains(t, output, "#FF5733")
}

func TestColumnCreate_DefaultColor(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "Review"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "column created")
	assert.Contains(t, output, "Review")
	assert.Contains(t, output, "#60A5FA")
}

func TestColumnCreate_WithWIPLimit(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "Review", "--wip-limit", "5"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "WIP Limit")
	assert.Contains(t, output, "5")
}

func TestColumnCreate_MissingName(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestColumnCreate_MissingBoardScope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--name", "Review"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestColumnCreate_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("color", "", "")
	cmd.Flags().Int("wip-limit", 0, "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "Review", "--color", "#FF5733", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"column"`)
	assert.Contains(t, output, "Review")
	assert.Contains(t, output, "#FF5733")
}

func TestColumnReorder_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt2.Close()

	ids := make([]string, len(columns))
	for i, c := range columns {
		ids[i] = c.ID
	}
	// reverse
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}

	args := []string{"--db-path", dbPath, "--board-id", setup.Board.ID}
	for _, id := range ids {
		args = append(args, "--column-id", id)
	}

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().StringArray("column-id", nil, "")
	require.NoError(t, cmd.ParseFlags(args))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnReorderWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Columns reordered")

	rt3, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	reordered, err := rt3.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt3.Close()

	for i, col := range reordered {
		assert.Equal(t, ids[i], col.ID, "position %d", i+1)
	}
}

func TestColumnReorder_MissingColumnID(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt2.Close()

	ids := []string{columns[1].ID, columns[2].ID}

	args := []string{"--db-path", dbPath, "--board-id", setup.Board.ID}
	for _, id := range ids {
		args = append(args, "--column-id", id)
	}

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().StringArray("column-id", nil, "")
	require.NoError(t, cmd.ParseFlags(args))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnReorderWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not included")
}

func TestColumnReorder_MissingBoardScope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().StringArray("column-id", nil, "")
	cmd.Flags().String("board-id", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--column-id", "c1"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnReorderWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "board")
}

func TestColumnReorder_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt2.Close()

	ids := make([]string, len(columns))
	for i, c := range columns {
		ids[i] = c.ID
	}

	args := []string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--json"}
	for _, id := range ids {
		args = append(args, "--column-id", id)
	}

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().StringArray("column-id", nil, "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags(args))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnReorderWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"reorder"`)
	assert.Contains(t, output, "board_id")
}

// -- Column Delete Tests --

func TestColumnDelete_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt2.Close()

	store := state.NewStore(dir + "/state.json")
	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("column-id", columns[0].ID, "")
	cmd.Flags().String("column", "", "")
	cmd.Flags().String("move-tasks-to", "", "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--column-id", columns[0].ID, "--yes"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnDeleteWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "column deleted")

	rt3, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	remaining, err := rt3.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt3.Close()

	assert.Len(t, remaining, 2)
}

func TestColumnDelete_WithTasksRequiresMove(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)

	// Create a task in the first column.
	_, err = rt2.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &columns[0].ID,
		Title:       "Task in column",
	})
	require.NoError(t, err)
	rt2.Close()

	store := state.NewStore(dir + "/state.json")
	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("column-id", columns[0].ID, "")
	cmd.Flags().String("column", "", "")
	cmd.Flags().String("move-tasks-to", "", "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--column-id", columns[0].ID, "--yes"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnDeleteWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "move-tasks-to")
}

func TestColumnDelete_WithReassign(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)

	// Create a task in the first column.
	task, err := rt2.TaskService.CreateTask(context.Background(), application.CreateTaskInput{
		ProviderID:  setup.Provider.ID,
		WorkspaceID: setup.Workspace.ID,
		BoardID:     &setup.Board.ID,
		ColumnID:    &columns[0].ID,
		Title:       "Task to move",
	})
	require.NoError(t, err)
	rt2.Close()

	store := state.NewStore(dir + "/state.json")
	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("column-id", columns[0].ID, "")
	cmd.Flags().String("column", "", "")
	cmd.Flags().String("move-tasks-to", columns[1].ID, "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--column-id", columns[0].ID, "--move-tasks-to", columns[1].ID, "--yes"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnDeleteWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "column deleted")

	rt3, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	remaining, err := rt3.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)

	// Verify task was moved.
	tasks, err := rt3.TaskFlow.ListTasks(context.Background(), application.ListTaskFilters{
		WorkspaceID: setup.Workspace.ID,
		ColumnID:    columns[1].ID,
	})
	require.NoError(t, err)
	rt3.Close()

	assert.Len(t, remaining, 2)
	assert.Len(t, tasks, 1)
	assert.Equal(t, task.ID, tasks[0].ID)
}

func TestColumnDelete_RequiresYes(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	columns, err := rt2.ContextService.ListColumns(context.Background(), setup.Board.ID)
	require.NoError(t, err)
	rt2.Close()

	store := state.NewStore(dir + "/state.json")
	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("column-id", columns[0].ID, "")
	cmd.Flags().String("column", "", "")
	cmd.Flags().String("move-tasks-to", "", "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--column-id", columns[0].ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnDeleteWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "yes")
}

func TestColumnDelete_NotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	store := state.NewStore(dir + "/state.json")
	require.NoError(t, store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID}))

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("board-id", setup.Board.ID, "")
	cmd.Flags().String("column-id", "invalid-id", "")
	cmd.Flags().String("column", "", "")
	cmd.Flags().String("move-tasks-to", "", "")
	cmd.Flags().Bool("yes", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--column-id", "invalid-id", "--yes"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnDeleteWithStore(cmd, ns, store)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
