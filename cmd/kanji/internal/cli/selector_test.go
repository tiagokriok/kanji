package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectorError_Error(t *testing.T) {
	err := &SelectorError{Code: "not_found", Message: "workspace not found"}
	assert.Equal(t, "workspace not found", err.Error())
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("workspace", "My Workspace")
	assert.Equal(t, "not_found", err.(*SelectorError).Code)
	assert.Contains(t, err.Error(), "workspace")
	assert.Contains(t, err.Error(), "My Workspace")
}

func TestNewAmbiguous(t *testing.T) {
	err := NewAmbiguous("board", "Main", 2)
	assert.Equal(t, "ambiguous", err.(*SelectorError).Code)
	assert.Contains(t, err.Error(), "board")
	assert.Contains(t, err.Error(), "Main")
	assert.Contains(t, err.Error(), "2")
}

func TestNewMismatch(t *testing.T) {
	err := NewMismatch("workspace", "expected-1", "got-2")
	assert.Equal(t, "mismatch", err.(*SelectorError).Code)
	assert.Contains(t, err.Error(), "workspace")
	assert.Contains(t, err.Error(), "expected-1")
	assert.Contains(t, err.Error(), "got-2")
}

func TestNewValidation(t *testing.T) {
	err := NewValidation("name is required")
	assert.Equal(t, "validation", err.(*SelectorError).Code)
	assert.Contains(t, err.Error(), "name is required")
}

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "hello", NormalizeName("hello"))
	assert.Equal(t, "hello", NormalizeName("  hello  "))
	assert.Equal(t, "hello world", NormalizeName("  Hello World  "))
	assert.Equal(t, "hello", NormalizeName("HELLO"))
}

func TestExactMatch(t *testing.T) {
	assert.True(t, ExactMatch("hello", "hello"))
	assert.True(t, ExactMatch("hello", "  hello  "))
	assert.True(t, ExactMatch("hello", "HELLO"))
	assert.False(t, ExactMatch("hello", "world"))
	assert.False(t, ExactMatch("hello", "hell"))
}

func TestSelectorError_Is(t *testing.T) {
	assert.True(t, errors.Is(NewNotFound("x", "y"), ErrNotFound))
	assert.True(t, errors.Is(NewAmbiguous("x", "y", 2), ErrAmbiguous))
	assert.True(t, errors.Is(NewMismatch("x", "a", "b"), ErrMismatch))
	assert.True(t, errors.Is(NewValidation("bad"), ErrValidation))

	assert.False(t, errors.Is(NewNotFound("x", "y"), ErrAmbiguous))
	assert.False(t, errors.Is(errors.New("other"), ErrNotFound))
}
