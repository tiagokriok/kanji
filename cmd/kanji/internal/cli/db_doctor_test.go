package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/state"
)

func TestDBDoctor_Ok(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Isolate state store.
	stateDir := filepath.Join(dir, "state")
	t.Setenv("XDG_CONFIG_HOME", stateDir)

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

	// Isolate state store.
	stateDir := filepath.Join(dir, "state")
	t.Setenv("XDG_CONFIG_HOME", stateDir)

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

	// Isolate state store.
	stateDir := filepath.Join(dir, "state")
	t.Setenv("XDG_CONFIG_HOME", stateDir)

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

func TestDBDoctor_DuplicateNames(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Isolate state store.
	stateDir := filepath.Join(dir, "state")
	t.Setenv("XDG_CONFIG_HOME", stateDir)

	// Migrate and bootstrap.
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

	// Drop uniqueness index and insert duplicate workspace via raw SQL.
	cfg, err := ResolveConfig(cmd2)
	require.NoError(t, err)
	adapter, err := db.NewSQLiteAdapter(cfg.DBPath)
	require.NoError(t, err)
	defer adapter.Close()

	_, err = adapter.Raw().Exec("DROP INDEX IF EXISTS idx_workspaces_name_unique")
	require.NoError(t, err)

	var providerID string
	err = adapter.Raw().QueryRow("SELECT provider_id FROM workspaces LIMIT 1").Scan(&providerID)
	require.NoError(t, err)

	_, err = adapter.Raw().Exec(
		"INSERT INTO workspaces (id, provider_id, remote_id, name) VALUES (?, ?, ?, ?)",
		"ws-dup", providerID, nil, "Default Workspace",
	)
	require.NoError(t, err)

	// Doctor.
	cmd3 := &cobra.Command{}
	cmd3.Flags().String("db-path", "", "")
	require.NoError(t, cmd3.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd3.SetOut(buf)

	err = runDBDoctor(cmd3, ns)
	require.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "duplicate_names")
	assert.Contains(t, output, "default workspace")
}

func TestDBDoctor_DanglingContext(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Migrate and bootstrap.
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

	// Set up state store with dangling reference.
	stateDir := filepath.Join(dir, "state")
	t.Setenv("XDG_CONFIG_HOME", stateDir)
	store, err := defaultStateStore()
	require.NoError(t, err)
	err = store.SetCLIContext(ns.Key, state.CLIContext{WorkspaceID: "nonexistent-ws-id"})
	require.NoError(t, err)

	// Doctor.
	cmd3 := &cobra.Command{}
	cmd3.Flags().String("db-path", "", "")
	require.NoError(t, cmd3.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd3.SetOut(buf)

	err = runDBDoctor(cmd3, ns)
	require.Error(t, err)

	output := buf.String()
	assert.Contains(t, output, "dangling_context")
	assert.Contains(t, output, "nonexistent-ws-id")
}
