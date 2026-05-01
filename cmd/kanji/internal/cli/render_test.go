package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTable(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderTable(buf, []string{"ID", "Name", "Status"}, [][]string{
		{"1", "Task One", "Todo"},
		{"2", "Task Two", "Doing"},
	})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Status")
	assert.Contains(t, output, "Task One")
	assert.Contains(t, output, "Task Two")
}

func TestRenderTable_EmptyRows(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderTable(buf, []string{"ID", "Name"}, [][]string{})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
}

func TestRenderKV(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderKV(buf, map[string]string{
		"ID":     "ws-1",
		"Name":   "Default Workspace",
		"Boards": "3",
	})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID:")
	assert.Contains(t, output, "ws-1")
	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "Default Workspace")
}

func TestRenderWrappedJSON(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderWrappedJSON(buf, "workspace", map[string]string{
		"id":   "ws-1",
		"name": "Default",
	})
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"workspace"`)
	assert.Contains(t, output, `"id"`)
	assert.Contains(t, output, `"name"`)
}

func TestRenderWrappedListJSON(t *testing.T) {
	buf := new(strings.Builder)
	items := []map[string]string{
		{"id": "ws-1", "name": "A"},
		{"id": "ws-2", "name": "B"},
	}
	err := RenderWrappedListJSON(buf, "workspaces", items, 2)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"workspaces"`)
	assert.Contains(t, output, `"count"`)
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "ws-1")
	assert.Contains(t, output, "ws-2")
}

func TestRenderJSONError(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderJSONError(buf, "not_found", "workspace not found")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"error"`)
	assert.Contains(t, output, `"code"`)
	assert.Contains(t, output, `"message"`)
	assert.Contains(t, output, "not_found")
	assert.Contains(t, output, "workspace not found")
}

func TestRenderJSONError_WithDetails(t *testing.T) {
	buf := new(strings.Builder)
	err := RenderJSONError(buf, "validation", "invalid input", "name is required")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"details"`)
	assert.Contains(t, output, "name is required")
}
