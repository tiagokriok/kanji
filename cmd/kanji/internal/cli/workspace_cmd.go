package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/domain"
)

func newWorkspaceCommand() *cobra.Command {
	ws := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace operations",
	}
	ws.AddCommand(newWorkspaceListCommand())
	ws.AddCommand(newWorkspaceGetCommand())
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

func newWorkspaceGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a workspace by ID or name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runWorkspaceGet(cmd, ns)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	return cmd
}

func runWorkspaceGet(cmd *cobra.Command, ns Namespace) error {
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

	ctx := context.Background()
	workspaces, err := rt.ContextService.ListWorkspaces(ctx)
	if err != nil {
		return err
	}

	var workspace domain.Workspace
	if cmd.Flags().Changed("workspace-id") {
		id, _ := cmd.Flags().GetString("workspace-id")
		found := false
		for _, ws := range workspaces {
			if ws.ID == id {
				workspace = ws
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", id)
		}
	} else if cmd.Flags().Changed("workspace") {
		name, _ := cmd.Flags().GetString("workspace")
		found := false
		for _, ws := range workspaces {
			if ExactMatch(ws.Name, name) {
				workspace = ws
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", name)
		}
	} else {
		return NewValidation("workspace-id or workspace is required")
	}

	if cfg.JSON {
		return RenderWrappedJSON(cmd.OutOrStdout(), "workspace", map[string]string{
			"id":   workspace.ID,
			"name": workspace.Name,
		})
	}

	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":   workspace.ID,
		"Name": workspace.Name,
	})
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
