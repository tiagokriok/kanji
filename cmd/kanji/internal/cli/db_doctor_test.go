package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBDoctor_Ok(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Bootstrap and migrate.
	cmd1 := &cobra.Command{}
	cmd1.Flags().String("db-path", "", "")
	require.NoError(t, cmd1.ParseFlags([]string{"--db-path", dbPath}))
	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateUp(cmd1, ns)
	require.NoError(t, err)

	cmd2 := &cobra.Command{}
	cmd2.Flags().String("db-path", "", "")
	require.NoError(t, cmd2.ParseFlags([]string{"--db-path", dbPath}))
	err = runDataBootstrap(cmd2, []string{})
	require.NoError(t, err)

	// Doctor.
	cmd3 := &cobra.Command{}
	cmd3.Flags().String("db-path", "", "")
	require.NoError(t, cmd3.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd3.SetOut(buf)

	err = runDBDoctor(cmd3, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "OK")
}

func TestDBDoctor_MissingBootstrap(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Migrate but don't bootstrap.
	cmd1 := &cobra.Command{}
	cmd1.Flags().String("db-path", "", "")
	require.NoError(t, cmd1.ParseFlags([]string{"--db-path", dbPath}))
	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBMigrateUp(cmd1, ns)
	require.NoError(t, err)

	// Doctor.
	cmd2 := &cobra.Command{}
	cmd2.Flags().String("db-path", "", "")
	require.NoError(t, cmd2.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd2.SetOut(buf)

	err = runDBDoctor(cmd2, ns)
	require.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "Found")
	assert.Contains(t, output, "bootstrap")
}

func TestDBDoctor_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBDoctor(cmd, ns)
	require.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "findings")
	assert.Contains(t, output, "status")
}
