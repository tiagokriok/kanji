package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
)

// RuntimeConfig holds the resolved CLI configuration after applying
// flag > env > default precedence.
type RuntimeConfig struct {
	DBPath  string
	JSON    bool
	Verbose bool
	Context string // from KANJI_CONTEXT env var; used by namespace resolution (P1-03).
}

// ResolveConfig extracts persistent flags and env vars into a RuntimeConfig.
// Precedence:
//
//	--db-path  > KANJI_DB_PATH > computed default
//	--json     > default false
//	--verbose  > default false
//	KANJI_CONTEXT (env only, no flag equivalent)
func ResolveConfig(cmd *cobra.Command) (RuntimeConfig, error) {
	var cfg RuntimeConfig

	if cmd.Flags().Changed("db-path") {
		path, _ := cmd.Flags().GetString("db-path")
		cfg.DBPath = path
	} else if v := os.Getenv("KANJI_DB_PATH"); v != "" {
		cfg.DBPath = v
	} else {
		defaultPath, err := db.DefaultDBPath(db.DefaultAppName)
		if err != nil {
			return RuntimeConfig{}, fmt.Errorf("resolve default db path: %w", err)
		}
		cfg.DBPath = defaultPath
	}

	cfg.JSON, _ = cmd.Flags().GetBool("json")
	cfg.Verbose, _ = cmd.Flags().GetBool("verbose")
	cfg.Context = os.Getenv("KANJI_CONTEXT")

	return cfg, nil
}
