package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePriority_Labels(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"critical", 0},
		{"urgent", 1},
		{"high", 2},
		{"medium", 3},
		{"low", 4},
		{"none", 5},
		{"CRITICAL", 0},
		{"Urgent", 1},
		{"  High  ", 2},
		{"MEDIUM", 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePriority(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParsePriority_Numeric(t *testing.T) {
	for i := 0; i <= 5; i++ {
		t.Run(string(rune('0'+i)), func(t *testing.T) {
			got, err := ParsePriority(string(rune('0' + i)))
			require.NoError(t, err)
			assert.Equal(t, i, got)
		})
	}
}

func TestParsePriority_Numeric_OutOfRange(t *testing.T) {
	tests := []string{"-1", "6", "10", "99"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParsePriority(input)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidation), "expected validation error")
			assert.Contains(t, err.Error(), "priority")
		})
	}
}

func TestParsePriority_InvalidInput(t *testing.T) {
	tests := []string{"", "foo", "important", "p1", "  "}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParsePriority(input)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidation), "expected validation error")
			assert.Contains(t, err.Error(), "priority")
		})
	}
}

func TestPriorityLabels(t *testing.T) {
	assert.Equal(t, []string{"critical", "urgent", "high", "medium", "low", "none"}, PriorityLabels())
}
