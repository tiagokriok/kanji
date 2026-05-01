package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
)

func TestResolveConfig_DBPath_Default(t *testing.T) {
	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	expected, err := db.DefaultDBPath(db.DefaultAppName)
	require.NoError(t, err)
	assert.Equal(t, expected, cfg.DBPath)
}

func TestResolveConfig_DBPath_EnvOverridesDefault(t *testing.T) {
	t.Setenv("KANJI_DB_PATH", "/tmp/env-kanji.db")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/env-kanji.db", cfg.DBPath)
}

func TestResolveConfig_DBPath_FlagOverridesEnv(t *testing.T) {
	t.Setenv("KANJI_DB_PATH", "/tmp/env-kanji.db")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", "/tmp/flag-kanji.db"}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.Equal(t, "/tmp/flag-kanji.db", cfg.DBPath)
}

func TestResolveConfig_JSON_Default(t *testing.T) {
	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.False(t, cfg.JSON)
}

func TestResolveConfig_JSON_Flag(t *testing.T) {
	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{"--json"}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.True(t, cfg.JSON)
}

func TestResolveConfig_Verbose_Default(t *testing.T) {
	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.False(t, cfg.Verbose)
}

func TestResolveConfig_Verbose_Flag(t *testing.T) {
	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{"--verbose"}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.True(t, cfg.Verbose)
}

func TestResolveConfig_Context_FromEnv(t *testing.T) {
	t.Setenv("KANJI_CONTEXT", "my-namespace")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.Equal(t, "my-namespace", cfg.Context)
}

func TestResolveConfig_Context_DefaultEmpty(t *testing.T) {
	// Ensure env is not set.
	_ = os.Unsetenv("KANJI_CONTEXT")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	assert.Empty(t, cfg.Context)
}

func TestResolveConfig_AllDefaults(t *testing.T) {
	_ = os.Unsetenv("KANJI_DB_PATH")
	_ = os.Unsetenv("KANJI_CONTEXT")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	expected, err := db.DefaultDBPath(db.DefaultAppName)
	require.NoError(t, err)

	assert.Equal(t, expected, cfg.DBPath)
	assert.False(t, cfg.JSON)
	assert.False(t, cfg.Verbose)
	assert.Empty(t, cfg.Context)
}

func TestResolveConfig_DBPath_FlagEmptyStringUsesEnv(t *testing.T) {
	// Edge case: explicit --db-path "" should still be treated as explicitly set.
	// In practice this is unusual; we test the current contract.
	t.Setenv("KANJI_DB_PATH", "/tmp/env-kanji.db")

	cmd := NewRootCommand()
	require.NoError(t, cmd.ParseFlags([]string{"--db-path", ""}))

	cfg, err := ResolveConfig(cmd)
	require.NoError(t, err)

	// Explicit empty flag takes precedence over env.
	assert.Empty(t, cfg.DBPath)
}

func TestDefaultDBPath_UsesUserConfigDir(t *testing.T) {
	path, err := db.DefaultDBPath(db.DefaultAppName)
	require.NoError(t, err)

	cfgDir, err := os.UserConfigDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(cfgDir, db.DefaultAppName, "app.db"), path)
}
