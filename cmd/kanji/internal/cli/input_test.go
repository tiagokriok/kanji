package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTextInput_Inline(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")
	require.NoError(t, cmd.Flags().Set("description", "hello world"))

	got, err := ResolveTextInput(cmd, "description", "file", false, strings.NewReader(""))
	require.NoError(t, err)
	assert.Equal(t, "hello world", got)
}

func TestResolveTextInput_File(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "input.md")
	require.NoError(t, os.WriteFile(fpath, []byte("file content"), 0o644))

	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")
	require.NoError(t, cmd.Flags().Set("file", fpath))

	got, err := ResolveTextInput(cmd, "description", "file", false, strings.NewReader(""))
	require.NoError(t, err)
	assert.Equal(t, "file content", got)
}

func TestResolveTextInput_Stdin(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")
	require.NoError(t, cmd.Flags().Set("file", "-"))

	stdin := strings.NewReader("from stdin")
	got, err := ResolveTextInput(cmd, "description", "file", false, stdin)
	require.NoError(t, err)
	assert.Equal(t, "from stdin", got)
}

func TestResolveTextInput_BothFlagsError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")
	require.NoError(t, cmd.Flags().Set("description", "inline"))
	require.NoError(t, cmd.Flags().Set("file", "some.txt"))

	_, err := ResolveTextInput(cmd, "description", "file", false, strings.NewReader(""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestResolveTextInput_NeitherFlagError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")

	_, err := ResolveTextInput(cmd, "description", "file", false, strings.NewReader(""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestResolveTextInput_AllowsEmpty(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")

	got, err := ResolveTextInput(cmd, "description", "file", true, strings.NewReader(""))
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestResolveTextInput_FileNotFound(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("description", "", "")
	cmd.Flags().String("file", "", "")
	require.NoError(t, cmd.Flags().Set("file", "/nonexistent/path.txt"))

	_, err := ResolveTextInput(cmd, "description", "file", false, strings.NewReader(""))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file")
}
