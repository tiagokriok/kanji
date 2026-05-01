package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuardBootstrap_PassesWhenBootstrapped(t *testing.T) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}

	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt.Close()

	_, err = rt.BootstrapService.EnsureDefaultSetup(context.Background())
	require.NoError(t, err)

	err = GuardBootstrap(rt)
	assert.NoError(t, err)
}

func TestGuardBootstrap_FailsWhenNotBootstrapped(t *testing.T) {
	dir := t.TempDir()
	cfg := RuntimeConfig{DBPath: filepath.Join(dir, "test.db")}

	rt, err := NewRuntime(context.Background(), cfg)
	require.NoError(t, err)
	defer rt.Close()

	err = GuardBootstrap(rt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kanji data bootstrap")
	assert.Contains(t, err.Error(), "not initialized")
}
