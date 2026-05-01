package cli

import (
	"time"
)

const dateLayout = "2006-01-02"

// ParseDueDate parses a due-date input into a time.Time value.
//
// Supported formats:
//   - YYYY-MM-DD       → resolved to end-of-day 23:59:59 UTC
//   - RFC3339          → parsed directly (e.g. 2025-12-25T14:30:00Z)
//
// Invalid inputs return a validation error with a clear message.
func ParseDueDate(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, NewValidation("due date is required; use YYYY-MM-DD or RFC3339")
	}

	// Attempt RFC3339 first since it is unambiguous.
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	// Attempt narrow date layout.
	if t, err := time.Parse(dateLayout, input); err == nil {
		return t.Add(23*time.Hour + 59*time.Minute + 59*time.Second).UTC(), nil
	}

	return time.Time{}, NewValidation("invalid due date " + input + "; use YYYY-MM-DD or RFC3339")
}
