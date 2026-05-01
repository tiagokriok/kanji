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

func TestDataSeedCmd_FirstRun(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Bootstrap first.
	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	// Seed.
	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runDataSeed(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Seed")
	assert.Contains(t, output, "completed")

	// Verify data was created.
	rt2, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt2.Close()

	workspaces, err := rt2.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(workspaces), 1)
}

func TestDataSeedCmd_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Bootstrap first.
	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	ns := Namespace{Key: "test-ns", Source: "cwd"}

	// First seed.
	cmd1 := &cobra.Command{}
	cmd1.Flags().String("db-path", "", "")
	require.NoError(t, cmd1.ParseFlags([]string{"--db-path", dbPath}))
	err = runDataSeed(cmd1, ns)
	require.NoError(t, err)

	// Second seed should also succeed.
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("db-path", "", "")
	require.NoError(t, cmd2.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd2.SetOut(buf)
	err = runDataSeed(cmd2, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Seed")
}

func TestDataSeedCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Bootstrap first.
	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runDataSeed(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "seed")
	assert.Contains(t, output, "completed")
}
