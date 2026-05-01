package cli

import (
	"fmt"
	"strings"
)

// SelectorError is a typed error for selector resolution failures.
type SelectorError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *SelectorError) Error() string {
	return e.Message
}

// Is reports whether target matches this error by code.
func (e *SelectorError) Is(target error) bool {
	t, ok := target.(*SelectorError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Sentinel errors for selector resolution.
var (
	ErrNotFound   = &SelectorError{Code: "not_found"}
	ErrAmbiguous  = &SelectorError{Code: "ambiguous"}
	ErrMismatch   = &SelectorError{Code: "mismatch"}
	ErrValidation = &SelectorError{Code: "validation"}
)

// NewNotFound returns a selector error for a missing resource.
func NewNotFound(resource, value string) error {
	return &SelectorError{
		Code:    "not_found",
		Message: fmt.Sprintf("%s not found: %s", resource, value),
	}
}

// NewAmbiguous returns a selector error for ambiguous matches.
func NewAmbiguous(resource, value string, matches int) error {
	return &SelectorError{
		Code:    "ambiguous",
		Message: fmt.Sprintf("%s selector is ambiguous: %q matches %d resources", resource, value, matches),
	}
}

// NewMismatch returns a selector error for scope or value mismatches.
func NewMismatch(resource, expected, got string) error {
	return &SelectorError{
		Code:    "mismatch",
		Message: fmt.Sprintf("%s mismatch: expected %q, got %q", resource, expected, got),
	}
}

// NewValidation returns a selector error for invalid input.
func NewValidation(message string) error {
	return &SelectorError{
		Code:    "validation",
		Message: message,
	}
}

// NormalizeName prepares a name for exact normalized matching:
// trims surrounding whitespace and converts to lowercase.
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ExactMatch reports whether a and b are equal after normalization.
func ExactMatch(a, b string) bool {
	return NormalizeName(a) == NormalizeName(b)
}
