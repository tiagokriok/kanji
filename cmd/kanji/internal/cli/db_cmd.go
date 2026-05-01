package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDBCommand() *cobra.Command {
	db := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
	}
	db.AddCommand(newDBInfoCommand())
	return db
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
