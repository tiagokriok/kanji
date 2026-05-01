package cli

import (
	"fmt"
	"regexp"
	"strings"
)

// ColumnSpec represents a parsed board column specification.
type ColumnSpec struct {
	Name  string
	Color string
}

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// ParseColumnSpec parses a single column spec in the form "Name:#RRGGBB".
// The color portion is optional; if omitted, Color is an empty string.
func ParseColumnSpec(raw string) (ColumnSpec, error) {
	name, color, found := strings.Cut(raw, ":")
	name = strings.TrimSpace(name)

	if name == "" {
		return ColumnSpec{}, fmt.Errorf("column name is required")
	}

	if found {
		color = strings.TrimSpace(color)
		if color != "" && !hexColorPattern.MatchString(color) {
			return ColumnSpec{}, fmt.Errorf("invalid color %q: expected hex format #RRGGBB", color)
		}
	}

	return ColumnSpec{Name: name, Color: color}, nil
}

// ParseColumnSpecs parses a slice of raw column specs, preserving order.
// Returns on the first invalid spec with an actionable error.
func ParseColumnSpecs(raws []string) ([]ColumnSpec, error) {
	result := make([]ColumnSpec, 0, len(raws))
	for _, raw := range raws {
		spec, err := ParseColumnSpec(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid column spec %q: %w", raw, err)
		}
		result = append(result, spec)
	}
	return result, nil
}
