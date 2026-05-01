package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// RenderTable writes a human-readable table with headers and rows.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return nil
	}

	// Compute column widths.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Helper to pad a string to a width.
	pad := func(s string, width int) string {
		if len(s) >= width {
			return s
		}
		return s + strings.Repeat(" ", width-len(s))
	}

	// Write header.
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, pad(h, widths[i]))
	}
	fmt.Fprintln(w)

	// Write separator.
	for i := range headers {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, strings.Repeat("-", widths[i]))
	}
	fmt.Fprintln(w)

	// Write rows.
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			fmt.Fprint(w, pad(cell, widths[i]))
		}
		fmt.Fprintln(w)
	}

	return nil
}

// RenderKV writes a human-readable key/value block.
func RenderKV(w io.Writer, pairs map[string]string) error {
	// Compute key width for alignment.
	maxKey := 0
	for k := range pairs {
		if len(k) > maxKey {
			maxKey = len(k)
		}
	}

	for k, v := range pairs {
		fmt.Fprintf(w, "%s:  %s\n", strings.Repeat(" ", maxKey-len(k))+k, v)
	}

	return nil
}

// RenderWrappedJSON writes a wrapped JSON success payload.
func RenderWrappedJSON(w io.Writer, key string, data interface{}) error {
	wrapper := map[string]interface{}{
		key: data,
	}
	return encodeJSON(w, wrapper)
}

// RenderWrappedListJSON writes a wrapped JSON list payload with count.
func RenderWrappedListJSON(w io.Writer, key string, items interface{}, count int) error {
	wrapper := map[string]interface{}{
		key:     items,
		"count": count,
	}
	return encodeJSON(w, wrapper)
}

// RenderJSONError writes a wrapped JSON error payload.
func RenderJSONError(w io.Writer, code, message string, details ...string) error {
	errObj := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if len(details) > 0 {
		errObj["details"] = details
	}
	wrapper := map[string]interface{}{
		"error": errObj,
	}
	return encodeJSON(w, wrapper)
}

func encodeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
