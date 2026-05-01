package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBMigrateStatus_Unmigrated(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateStatus(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "unmigrated")
}

func TestDBMigrateStatus_Migrated(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Run migrations first.
	cmd1 := &cobra.Command{}
	cmd1.Flags().String("db-path", "", "")
	require.NoError(t, cmd1.ParseFlags([]string{"--db-path", dbPath}))
	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateUp(cmd1, ns)
	require.NoError(t, err)

	// Check status.
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("db-path", "", "")
	require.NoError(t, cmd2.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd2.SetOut(buf)

	err = runDBMigrateStatus(cmd2, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "migrated")
	assert.Contains(t, output, "Version")
}

func TestDBMigrateStatus_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateStatus(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "version")
}
