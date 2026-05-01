package cli

import (
	"strconv"
	"strings"
)

// priorityLabelMap maps canonical priority labels to their numeric values.
var priorityLabelMap = map[string]int{
	"critical": 0,
	"urgent":   1,
	"high":     2,
	"medium":   3,
	"low":      4,
	"none":     5,
}

// ParsePriority normalizes a priority input to a canonical internal value (0-5).
// It accepts both numeric strings ("0" through "5") and case-insensitive labels
// such as "critical", "urgent", "high", "medium", "low", and "none".
// Invalid inputs return a validation error with an actionable message.
func ParsePriority(input string) (int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, NewValidation("priority is required; use a number 0-5 or a label: critical, urgent, high, medium, low, none")
	}

	// Try label match first (case-insensitive).
	if v, ok := priorityLabelMap[strings.ToLower(trimmed)]; ok {
		return v, nil
	}

	// Try numeric parse.
	if n, err := strconv.Atoi(trimmed); err == nil {
		if n >= 0 && n <= 5 {
			return n, nil
		}
		return 0, NewValidation("priority must be between 0 and 5; got " + trimmed)
	}

	return 0, NewValidation("invalid priority " + trimmed + "; use a number 0-5 or a label: critical, urgent, high, medium, low, none")
}

// PriorityLabels returns the sorted list of supported priority label strings.
func PriorityLabels() []string {
	return []string{"critical", "urgent", "high", "medium", "low", "none"}
}
