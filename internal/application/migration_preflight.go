package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

// PreflightMigrations runs duplicate-name diagnostics before applying
// uniqueness migrations. If any duplicates are found it returns an error
// listing each conflict with an actionable message.
func PreflightMigrations(ctx context.Context, db *sql.DB, setupRepo domain.SetupRepository) error {
	var messages []string

	wsDups, err := FindDuplicateWorkspaceNames(ctx, setupRepo)
	if err != nil {
		return fmt.Errorf("check workspace names: %w", err)
	}
	for _, d := range wsDups {
		messages = append(messages, fmt.Sprintf(`workspace name %q has %d duplicates: run "kanji db doctor" to see details`, d.Name, d.Count))
	}

	boardDups, err := FindDuplicateBoardNames(ctx, setupRepo)
	if err != nil {
		return fmt.Errorf("check board names: %w", err)
	}
	for _, d := range boardDups {
		messages = append(messages, fmt.Sprintf(`board name %q has %d duplicates: run "kanji db doctor" to see details`, d.Name, d.Count))
	}

	colDups, err := FindDuplicateColumnNames(ctx, setupRepo)
	if err != nil {
		return fmt.Errorf("check column names: %w", err)
	}
	for _, d := range colDups {
		messages = append(messages, fmt.Sprintf(`column name %q has %d duplicates: run "kanji db doctor" to see details`, d.Name, d.Count))
	}

	if len(messages) > 0 {
		return fmt.Errorf("migration preflight failed:\n%s", strings.Join(messages, "\n"))
	}
	return nil
}
