package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTUICommand_Exists(t *testing.T) {
	root := NewRootCommand()
	cmd, _, err := root.Find([]string{"tui"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "tui", cmd.Name())
}
