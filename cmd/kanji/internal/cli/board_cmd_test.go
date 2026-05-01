package cli

import (
	"context"
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
