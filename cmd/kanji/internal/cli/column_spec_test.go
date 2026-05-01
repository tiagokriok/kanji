package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseColumnSpec_Valid(t *testing.T) {
	spec, err := ParseColumnSpec("Todo:#60A5FA")
	require.NoError(t, err)
	assert.Equal(t, "Todo", spec.Name)
	assert.Equal(t, "#60A5FA", spec.Color)
}

func TestParseColumnSpec_MissingColor(t *testing.T) {
	spec, err := ParseColumnSpec("Backlog")
	require.NoError(t, err)
	assert.Equal(t, "Backlog", spec.Name)
	assert.Equal(t, "", spec.Color)
}

func TestParseColumnSpec_InvalidColor(t *testing.T) {
	_, err := ParseColumnSpec("Todo:#ZZZZZZ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "color")
}

func TestParseColumnSpec_EmptyName(t *testing.T) {
	_, err := ParseColumnSpec(":#60A5FA")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParseColumnSpec_EmptyNameNoColor(t *testing.T) {
	_, err := ParseColumnSpec("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParseColumnSpecs_Valid(t *testing.T) {
	specs, err := ParseColumnSpecs([]string{"Todo:#60A5FA", "Doing", "Done:#22C55E"})
	require.NoError(t, err)
	require.Len(t, specs, 3)
	assert.Equal(t, "Todo", specs[0].Name)
	assert.Equal(t, "#60A5FA", specs[0].Color)
	assert.Equal(t, "Doing", specs[1].Name)
	assert.Equal(t, "", specs[1].Color)
	assert.Equal(t, "Done", specs[2].Name)
	assert.Equal(t, "#22C55E", specs[2].Color)
}

func TestParseColumnSpecs_Invalid(t *testing.T) {
	_, err := ParseColumnSpecs([]string{"Todo:#60A5FA", "Review:#ZZZZZZ"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "color")
}
