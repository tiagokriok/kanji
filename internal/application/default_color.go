package application

import "github.com/tiagokriok/kanji/internal/domain"

// defaultColorPalette is the canonical board column color palette.
// Kept in sync with defaultColumnSpecs.
var defaultColorPalette = []string{
	"#60A5FA",
	"#F59E0B",
	"#22C55E",
}

// fallbackColor is returned when the palette is empty.
const fallbackColor = "#9CA3AF"

// NextDefaultColor returns the next palette color for a board given its
// existing columns. The color is selected using modulo over the palette so
// the sequence cycles deterministically.
func NextDefaultColor(columns []domain.Column) string {
	if len(defaultColorPalette) == 0 {
		return fallbackColor
	}
	idx := len(columns) % len(defaultColorPalette)
	return defaultColorPalette[idx]
}
