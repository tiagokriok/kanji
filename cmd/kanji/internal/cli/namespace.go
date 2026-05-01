package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Namespace holds the resolved namespace key and its source.
type Namespace struct {
	Key    string // the namespace identifier (cwd path or env override)
	Source string // "cwd" or "env"
}

// ResolveNamespace returns the active namespace.
//
// Precedence:
//  1. KANJI_CONTEXT env var (source: "env")
//  2. Current working directory, cleaned with filepath.Clean (source: "cwd")
func ResolveNamespace() (Namespace, error) {
	if v := os.Getenv("KANJI_CONTEXT"); v != "" {
		return Namespace{Key: v, Source: "env"}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return Namespace{}, fmt.Errorf("get working directory: %w", err)
	}

	return Namespace{Key: filepath.Clean(cwd), Source: "cwd"}, nil
}
