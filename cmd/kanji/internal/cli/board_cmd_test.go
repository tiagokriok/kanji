package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/state"
)

func TestBoardList_ExplicitScope(t *testing.T) {
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
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, setup.Board.Name)
}

func TestBoardList_MissingScope(t *testing.T) {
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
	err = runBoardList(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
}

func TestBoardList_InferredScope(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	// Set context workspace.
	store := state.NewStore(dir + "/state.json")
	_ = store.SetCLIContext("test-ns", state.CLIContext{WorkspaceID: setup.Workspace.ID})

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardListWithStore(cmd, ns, store)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, setup.Board.Name)
}

func TestBoardList_JSON(t *testing.T) {
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
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "boards")
	assert.Contains(t, output, setup.Board.Name)
}

// -- Board Create Tests --

func TestBoardCreate_MissingName(t *testing.T) {
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
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestBoardCreate_WithDefaults(t *testing.T) {
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
	cmd.Flags().String("name", "New Board", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--name", "New Board"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Board created")
	assert.Contains(t, output, "New Board")

	// Verify default columns were created.
	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt2.Close()
	boards, err := rt2.ContextService.ListBoards(context.Background(), setup.Workspace.ID)
	require.NoError(t, err)
	var createdBoardID string
	for _, b := range boards {
		if b.Name == "New Board" {
			createdBoardID = b.ID
			break
		}
	}
	require.NotEmpty(t, createdBoardID)
	columns, err := rt2.ContextService.ListColumns(context.Background(), createdBoardID)
	require.NoError(t, err)
	require.Len(t, columns, 3)
}

func TestBoardCreate_WithCustomColumns(t *testing.T) {
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
	cmd.Flags().String("name", "Custom Board", "")
	cmd.Flags().StringArray("column", nil, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--name", "Custom Board", "--column", "Backlog:#FF0000", "--column", "In Progress:#00FF00"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardCreate(cmd, ns)
	require.NoError(t, err)

	// Verify only custom columns were created.
	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt2.Close()
	boards, err := rt2.ContextService.ListBoards(context.Background(), setup.Workspace.ID)
	require.NoError(t, err)
	var createdBoardID string
	for _, b := range boards {
		if b.Name == "Custom Board" {
			createdBoardID = b.ID
			break
		}
	}
	require.NotEmpty(t, createdBoardID)
	columns, err := rt2.ContextService.ListColumns(context.Background(), createdBoardID)
	require.NoError(t, err)
	require.Len(t, columns, 2)
	assert.Equal(t, "Backlog", columns[0].Name)
	assert.Equal(t, "#FF0000", columns[0].Color)
	assert.Equal(t, "In Progress", columns[1].Name)
	assert.Equal(t, "#00FF00", columns[1].Color)
}

func TestBoardCreate_SetContext(t *testing.T) {
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
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("name", "Context Board", "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--name", "Context Board", "--set-context"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext("test-ns")
	require.NoError(t, err)
	assert.Equal(t, setup.Workspace.ID, ctx.WorkspaceID)
	assert.NotEmpty(t, ctx.BoardID)
}

func TestBoardCreate_JSON(t *testing.T) {
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
	cmd.Flags().String("name", "JSON Board", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--name", "JSON Board", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &payload))
	board, ok := payload["board"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "JSON Board", board["name"])
	assert.NotEmpty(t, board["id"])
}

// -- Board Update Tests --

func TestBoardUpdate_ByBoardID(t *testing.T) {
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
	cmd.Flags().String("name", "Renamed Board", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "Renamed Board"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Board updated")
	assert.Contains(t, output, "Renamed Board")
}

func TestBoardUpdate_ByBoardName(t *testing.T) {
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
	cmd.Flags().String("board", setup.Board.Name, "")
	cmd.Flags().String("workspace-id", setup.Workspace.ID, "")
	cmd.Flags().String("name", "Renamed by Name", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board", setup.Board.Name, "--workspace-id", setup.Workspace.ID, "--name", "Renamed by Name"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Board updated")
	assert.Contains(t, output, "Renamed by Name")
}

func TestBoardUpdate_MissingName(t *testing.T) {
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
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardUpdate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestBoardUpdate_JSON(t *testing.T) {
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
	cmd.Flags().String("name", "JSON Rename", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", setup.Board.ID, "--name", "JSON Rename", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(output), &payload))
	board, ok := payload["board"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "JSON Rename", board["name"])
	assert.Equal(t, setup.Board.ID, board["id"])
}
