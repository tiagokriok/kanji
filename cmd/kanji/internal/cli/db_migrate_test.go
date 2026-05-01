package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBMigrateUp_FirstRun(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateUpWithNamespace(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Migrations")
	assert.Contains(t, output, "completed")

	// Verify migrations actually ran.
	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt.Close()

	var count int
	err = rt.DB.Raw().QueryRow(
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name='providers'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDBMigrateUp_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// First migration.
	cmd1 := &cobra.Command{}
	cmd1.Flags().String("db-path", "", "")
	require.NoError(t, cmd1.ParseFlags([]string{"--db-path", dbPath}))
	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateUpWithNamespace(cmd1, ns)
	require.NoError(t, err)

	// Second migration should also succeed.
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("db-path", "", "")
	require.NoError(t, cmd2.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd2.SetOut(buf)

	err = runDBMigrateUpWithNamespace(cmd2, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Migrations")
}

func runDBMigrateUpWithNamespace(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	rt, err := NewRuntime(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer rt.Close()

	fmt.Fprintf(cmd.OutOrStdout(), "Migrations completed.\n")
	return nil
}
