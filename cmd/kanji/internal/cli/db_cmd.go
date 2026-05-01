package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
)

func newDBCommand() *cobra.Command {
	db := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
	}
	db.AddCommand(newDBInfoCommand())
	db.AddCommand(newDBMigrateCommand())
	return db
}

func newDBMigrateCommand() *cobra.Command {
	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Database migrations",
	}
	migrate.AddCommand(newDBMigrateUpCommand())
	migrate.AddCommand(newDBMigrateStatusCommand())
	return migrate
}

func newDBMigrateUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run forward migrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runDBMigrateUp(cmd, ns)
		},
	}
}

func newDBMigrateStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runDBMigrateStatus(cmd, ns)
		},
	}
}

func runDBMigrateStatus(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	adapter, err := db.NewSQLiteAdapter(cfg.DBPath)
	if err != nil {
		return err
	}
	defer adapter.Close()

	var version int64
	var status string
	err = adapter.Raw().QueryRow(
		"SELECT version_id FROM goose_db_version ORDER BY version_id DESC LIMIT 1",
	).Scan(&version)
	if err != nil {
		status = "unmigrated"
		version = 0
	} else {
		status = "migrated"
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"status":  status,
			"version": version,
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "migrate", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Status:  %s\n", status)
	fmt.Fprintf(cmd.OutOrStdout(), "Version: %d\n", version)
	return nil
}

func runDBMigrateUp(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	rt, err := NewRuntime(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer rt.Close()

	fmt.Fprintln(cmd.OutOrStdout(), "Migrations completed.")
	return nil
}

func newDBInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show database status and metadata",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runDBInfo(cmd, ns)
		},
	}
}

func runDBInfo(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	exists := "no"
	if _, err := os.Stat(cfg.DBPath); err == nil {
		exists = "yes"
	}

	bootstrapped := "no"
	if exists == "yes" {
		rt, err := NewRuntime(context.Background(), cfg)
		if err == nil {
			defer rt.Close()
			workspaces, _ := rt.ContextService.ListWorkspaces(context.Background())
			if len(workspaces) > 0 {
				bootstrapped = "yes"
			}
		}
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"db_path":      cfg.DBPath,
			"exists":       exists == "yes",
			"namespace":    ns.Key,
			"bootstrapped": bootstrapped == "yes",
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "db", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "DB Path:      %s\n", cfg.DBPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Exists:       %s\n", exists)
	fmt.Fprintf(cmd.OutOrStdout(), "Namespace:    %s\n", ns.Key)
	fmt.Fprintf(cmd.OutOrStdout(), "Bootstrapped: %s\n", bootstrapped)
	return nil
}
