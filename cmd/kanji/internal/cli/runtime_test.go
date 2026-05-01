package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntime(t *testing.T) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}

	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt.Close()

	assert.NotNil(t, rt.DB)
	assert.NotNil(t, rt.Store)
	assert.NotNil(t, rt.BootstrapService)
	assert.NotNil(t, rt.TaskService)
	assert.NotNil(t, rt.TaskFlow)
	assert.NotNil(t, rt.CommentService)
	assert.NotNil(t, rt.ContextService)
}

func TestNewRuntime_MigrationsRun(t *testing.T) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}

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

func TestNewRuntime_Close(t *testing.T) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}

	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)

	err = rt.Close()
	assert.NoError(t, err)
}

func TestNewRuntime_InvalidDBPath(t *testing.T) {
	// Use a path in a non-existent directory that cannot be created.
	cfg := RuntimeConfig{DBPath: "/dev/null/invalid/test.db"}

	_, err := NewRuntime(context.Background(), cfg)
	require.Error(t, err)
}
