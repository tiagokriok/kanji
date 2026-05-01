package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newWorkspaceCommand() *cobra.Command {
	ws := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace operations",
	}
	ws.AddCommand(newWorkspaceListCommand())
	return ws
}

func newWorkspaceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runWorkspaceList(cmd, ns)
		},
	}
}

func runWorkspaceList(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	rt, err := NewRuntime(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer rt.Close()

	if err := GuardBootstrap(rt); err != nil {
		return err
	}

	workspaces, err := rt.ContextService.ListWorkspaces(context.Background())
	if err != nil {
		return err
	}

	if cfg.JSON {
		items := make([]map[string]string, len(workspaces))
		for i, ws := range workspaces {
			items[i] = map[string]string{
				"id":   ws.ID,
				"name": ws.Name,
			}
		}
		return RenderWrappedListJSON(cmd.OutOrStdout(), "workspaces", items, len(workspaces))
	}

	headers := []string{"ID", "Name"}
	rows := make([][]string, len(workspaces))
	for i, ws := range workspaces {
		rows[i] = []string{ws.ID, ws.Name}
	}
	return RenderTable(cmd.OutOrStdout(), headers, rows)
}
