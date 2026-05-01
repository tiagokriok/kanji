package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoardGet_ByID(t *testing.T) {
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
	err = runBoardGet(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, setup.Board.Name)
	assert.Contains(t, output, setup.Board.ID)
}

func TestBoardGet_ByName(t *testing.T) {
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
	cmd.Flags().String("board", setup.Board.Name, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--workspace-id", setup.Workspace.ID, "--board", setup.Board.Name}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardGet(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, setup.Board.Name)
}

func TestBoardGet_NotFound(t *testing.T) {
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
	cmd.Flags().String("board-id", "invalid", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--board-id", "invalid"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runBoardGet(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBoardGet_JSON(t *testing.T) {
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
	err = runBoardGet(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "board")
	assert.Contains(t, output, setup.Board.Name)
}
