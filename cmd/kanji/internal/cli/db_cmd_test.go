package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBInfo_MissingDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nonexistent", "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBInfoWithNamespace(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, dbPath)
	assert.Contains(t, output, "Exists")
	assert.Contains(t, output, "no")
}

func TestDBInfo_ExistingNotBootstrapped(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Create empty DB by running migrations without bootstrap.
	cfg := RuntimeConfig{DBPath: dbPath}
	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	rt.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err = runDBInfoWithNamespace(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, dbPath)
	assert.Contains(t, output, "yes")
	assert.Contains(t, output, "Bootstrapped")
	assert.Contains(t, output, "no")
}

func TestDBInfo_Bootstrapped(t *testing.T) {
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
	err = runDBInfoWithNamespace(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, dbPath)
	assert.Contains(t, output, "yes")
	assert.Contains(t, output, "Bootstrapped")
	assert.Contains(t, output, "yes")
}

func TestDBInfo_JSON(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	cmd := &cobra.Command{}
	cmd.Flags().String("db-path", "", "")
	cmd.Flags().Bool("json", false, "")
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", dbPath, "--json"}))
	buf := new(strings.Builder)
	cmd.SetOut(buf)

	ns := Namespace{Key: "test-ns", Source: "cwd"}
	err := runDBInfoWithNamespace(cmd, ns)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "db_path")
	assert.Contains(t, output, "namespace")
	assert.Contains(t, output, "bootstrapped")
}

func runDBInfoWithNamespace(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	exists := "no"
	if _, err := os.Stat(cfg.DBPath); err == nil {
		exists = "yes"
	}

	bootstrapped := "no"
	if exists == "yes" {
		rt, err := NewRuntime(context.Background(), cfg)
		if err == nil {
			defer rt.Close()
			workspaces, _ := rt.ContextService.ListWorkspaces(context.Background())
			if len(workspaces) > 0 {
				bootstrapped = "yes"
			}
		}
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"db_path":      cfg.DBPath,
			"exists":       exists == "yes",
			"namespace":    ns.Key,
			"bootstrapped": bootstrapped == "yes",
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "db", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "DB Path:      %s\n", cfg.DBPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Exists:       %s\n", exists)
	fmt.Fprintf(cmd.OutOrStdout(), "Namespace:    %s\n", ns.Key)
	fmt.Fprintf(cmd.OutOrStdout(), "Bootstrapped: %s\n", bootstrapped)
	return nil
}
