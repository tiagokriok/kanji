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

func TestWorkspaceList_Empty(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Bootstrap but don't create extra workspaces.
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
	err = runWorkspaceList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Default Workspace")
}

func TestWorkspaceList_JSON(t *testing.T) {
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
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runWorkspaceList(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspaces")
	assert.Contains(t, output, "count")
	assert.Contains(t, output, "Default Workspace")
}

func TestWorkspaceCreate_Success(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--name", "Test Workspace"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace created")
	assert.Contains(t, output, "Test Workspace")
	assert.Contains(t, output, "Main")
}

func TestWorkspaceCreate_WithSetContext(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	store := state.NewStore(dir + "/state.json")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--name", "Test Workspace", "--set-context"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceCreateWithStore(cmd, ns, store)
	require.NoError(t, err)

	ctx, err := store.GetCLIContext(ns.Key)
	require.NoError(t, err)
	assert.NotEmpty(t, ctx.WorkspaceID)
	assert.NotEmpty(t, ctx.BoardID)
}

func TestWorkspaceCreate_JSON(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--name", "Test Workspace", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceCreate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace")
	assert.Contains(t, output, "Test Workspace")
	assert.Contains(t, output, "Main")
}

func TestWorkspaceCreate_MissingName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db")}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestWorkspaceCreate_NotBootstrapped(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("set-context", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--name", "Test"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceCreate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestWorkspaceUpdate_ByID(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--workspace-id", ws[0].ID, "--name", "Renamed Workspace"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runWorkspaceUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace updated")
	assert.Contains(t, output, "Renamed Workspace")
}

func TestWorkspaceUpdate_ByName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--workspace", ws[0].Name, "--name", "Renamed Workspace"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runWorkspaceUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace updated")
	assert.Contains(t, output, "Renamed Workspace")
}

func TestWorkspaceUpdate_JSON(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--workspace-id", ws[0].ID, "--name", "Renamed", "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runWorkspaceUpdate(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "workspace")
	assert.Contains(t, output, "Renamed")
}

func TestWorkspaceUpdate_MissingName(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	ws, err := rt.ContextService.ListWorkspaces(context.Background())
	require.NoError(t, err)
	require.Len(t, ws, 1)

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--workspace-id", ws[0].ID}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runWorkspaceUpdate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestWorkspaceUpdate_NotFound(t *testing.T) {
	rt, dir := setupBootstrappedRuntime(t)
	defer rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().String("workspace-id", "", "")
	cmd.Flags().String("workspace", "", "")
	cmd.Flags().String("name", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", filepath.Join(dir, "test.db"), "--workspace-id", "invalid-id", "--name", "New Name"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runWorkspaceUpdate(cmd, ns)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
