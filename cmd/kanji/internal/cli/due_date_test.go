package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDueDate_YYYY_MM_DD(t *testing.T) {
	got, err := ParseDueDate("2025-12-25")
	require.NoError(t, err)

	want := time.Date(2025, 12, 25, 23, 59, 59, 0, time.UTC)
	assert.Equal(t, want, got)
	assert.Equal(t, time.UTC, got.Location())
}

func TestParseDueDate_RFC3339(t *testing.T) {
	input := "2025-12-25T14:30:00Z"
	got, err := ParseDueDate(input)
	require.NoError(t, err)

	want := time.Date(2025, 12, 25, 14, 30, 0, 0, time.UTC)
	assert.Equal(t, want, got)
	assert.Equal(t, time.UTC, got.Location())
}

func TestParseDueDate_RFC3339_WithOffset(t *testing.T) {
	input := "2025-12-25T14:30:00+02:00"
	got, err := ParseDueDate(input)
	require.NoError(t, err)

	want := time.Date(2025, 12, 25, 12, 30, 0, 0, time.UTC)
	assert.Equal(t, want, got.UTC())
}

func TestParseDueDate_InvalidInput(t *testing.T) {
	tests := []string{"", "foo", "12-25-2025", "25-12-2025", "2025/12/25"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDueDate(input)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidation), "expected validation error")
			assert.Contains(t, err.Error(), "due date")
		})
	}
}

func TestParseDueDate_LeapYear(t *testing.T) {
	got, err := ParseDueDate("2024-02-29")
	require.NoError(t, err)

	want := time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC)
	assert.Equal(t, want, got)
}

func TestParseDueDate_InvalidDate(t *testing.T) {
	_, err := ParseDueDate("2023-02-29")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrValidation), "expected validation error")
	assert.Contains(t, err.Error(), "due date")
}
