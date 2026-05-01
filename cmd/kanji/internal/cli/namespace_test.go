package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveNamespace_CwdDefault(t *testing.T) {
	// Ensure env is not set.
	t.Setenv("KANJI_CONTEXT", "")

	ns, err := ResolveNamespace()
	require.NoError(t, err)

	assert.Equal(t, "cwd", ns.Source)
	assert.NotEmpty(t, ns.Key)
	// Key should be a clean absolute path.
	assert.Equal(t, filepath.Clean(ns.Key), ns.Key)
}

func TestResolveNamespace_EnvOverride(t *testing.T) {
	t.Setenv("KANJI_CONTEXT", "my-custom-namespace")

	ns, err := ResolveNamespace()
	require.NoError(t, err)

	assert.Equal(t, "env", ns.Source)
	assert.Equal(t, "my-custom-namespace", ns.Key)
}

func TestResolveNamespace_EnvOverridesCwd(t *testing.T) {
	t.Setenv("KANJI_CONTEXT", "override-ns")

	ns, err := ResolveNamespace()
	require.NoError(t, err)

	assert.Equal(t, "env", ns.Source)
	assert.Equal(t, "override-ns", ns.Key)
	// Should not be the cwd.
	assert.NotContains(t, ns.Key, "/")
}

func TestResolveNamespace_CwdPathIsCleaned(t *testing.T) {
	// We cannot easily control os.Getwd in a subprocess test,
	// but we can verify that when ResolveNamespace returns a cwd-derived
	// key, it is already cleaned by filepath.Clean.
	t.Setenv("KANJI_CONTEXT", "")

	ns, err := ResolveNamespace()
	require.NoError(t, err)
	require.Equal(t, "cwd", ns.Source)

	cleaned := filepath.Clean(ns.Key)
	assert.Equal(t, cleaned, ns.Key, "namespace key should be a clean path")
}
