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

func TestColumnUpdate_Name(t *testing.T) {
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
	cmd.Flags().String("column-id", setup.Columns[0].ID, "")
	cmd.Flags().String("name", "Renamed", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--column-id", setup.Columns[0].ID, "--name", "Renamed"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Renamed")
}

func TestColumnUpdate_Color(t *testing.T) {
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
	cmd.Flags().String("column-id", setup.Columns[0].ID, "")
	cmd.Flags().String("color", "#FF0000", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--column-id", setup.Columns[0].ID, "--color", "#FF0000"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "#FF0000")
}

func TestColumnUpdate_ClearWIPLimit(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	setup, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	// Set a WIP limit first.
	_ = rt.ContextService.UpdateColumn(context.Background(), setup.Columns[0].ID, nil, nil, intPtr(5), false)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("column-id", setup.Columns[0].ID, "")
	cmd.Flags().Bool("clear-wip-limit", true, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--column-id", setup.Columns[0].ID, "--clear-wip-limit"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnUpdate(cmd, ns)
	require.NoError(t, err)
}

func TestColumnUpdate_JSON(t *testing.T) {
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
	cmd.Flags().String("column-id", setup.Columns[0].ID, "")
	cmd.Flags().String("name", "Renamed", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--column-id", setup.Columns[0].ID, "--name", "Renamed", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runColumnUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "column")
	assert.Contains(t, output, "Renamed")
}

func intPtr(v int) *int {
	return &v
}
