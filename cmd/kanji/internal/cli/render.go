package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
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

	// Sort keys for stable output.
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(w, "%s:  %s\n", strings.Repeat(" ", maxKey-len(k))+k, pairs[k])
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

// RenderWriteResult writes a human-readable success block for create/update
// operations. It prints the resource name, ID, and any additional fields as
// aligned key/value pairs.
func RenderWriteResult(w io.Writer, resourceName, id string, fields map[string]string) error {
	fmt.Fprintf(w, "%s created\n", resourceName)
	pairs := map[string]string{"ID": id}
	for k, v := range fields {
		pairs[k] = v
	}
	return RenderKV(w, pairs)
}

// RenderWriteResultJSON writes a wrapped JSON payload for write success.
func RenderWriteResultJSON(w io.Writer, resourceName string, data map[string]interface{}) error {
	return RenderWrappedJSON(w, resourceName, data)
}

// RenderWriteResolved writes an optional human metadata block showing context
// resolution sources. If resolved is empty, it writes nothing.
func RenderWriteResolved(w io.Writer, resolved map[string]string) error {
	if len(resolved) == 0 {
		return nil
	}
	fmt.Fprintln(w, "Resolved from:")
	return RenderKV(w, resolved)
}

// RenderDryRunImpact writes a human-readable dry-run impact summary.
func RenderDryRunImpact(w io.Writer, resourceName string, impact map[string]int) error {
	fmt.Fprintf(w, "Dry-run: %s delete impact\n", resourceName)
	pairs := map[string]string{}
	for k, v := range impact {
		pairs[k] = strconv.Itoa(v)
	}
	return RenderKV(w, pairs)
}

// RenderDryRunImpactJSON writes a JSON dry-run impact payload.
func RenderDryRunImpactJSON(w io.Writer, resourceName string, impact map[string]int) error {
	payload := map[string]interface{}{
		"dry_run": true,
		"impact":  impact,
	}
	return RenderWrappedJSON(w, resourceName, payload)
}

// RenderDeleteResult writes a human-readable delete success block.
func RenderDeleteResult(w io.Writer, resourceName, id string) error {
	if len(resourceName) > 0 {
		resourceName = strings.ToUpper(resourceName[:1]) + resourceName[1:]
	}
	fmt.Fprintf(w, "%s deleted\nID:  %s\n", resourceName, id)
	return nil
}

// RenderDeleteResultJSON writes a JSON delete success payload.
func RenderDeleteResultJSON(w io.Writer, resourceName string, id string, cascade bool) error {
	payload := map[string]interface{}{
		"id":      id,
		"deleted": true,
		"cascade": cascade,
	}
	return RenderWrappedJSON(w, resourceName, payload)
}

func encodeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
