package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/state"
)

func newWorkspaceCommand() *cobra.Command {
	ws := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace operations",
	}
	ws.AddCommand(newWorkspaceListCommand())
	ws.AddCommand(newWorkspaceGetCommand())
	ws.AddCommand(newWorkspaceCreateCommand())
	ws.AddCommand(newWorkspaceUpdateCommand())
	ws.AddCommand(newWorkspaceDeleteCommand())
	return ws
}

func newWorkspaceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runWorkspaceCreate(cmd, ns)
		},
	}
	cmd.Flags().String("name", "", "workspace name")
	cmd.Flags().Bool("set-context", false, "set context to the new workspace")
	return cmd
}

func newWorkspaceUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a workspace name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runWorkspaceUpdate(cmd, ns)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().String("name", "", "new workspace name")
	return cmd
}

func newWorkspaceDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workspace and its data",
		Long: `Delete a workspace and all its data (boards, columns, tasks, comments).

This is a destructive operation. Use --dry-run to preview the impact.
Requires both --yes and --cascade for actual deletion.`,
		Example: `  # Preview impact
  kanji workspace delete --workspace-id <id> --dry-run

  # Delete with confirmation
  kanji workspace delete --workspace-id <id> --yes --cascade`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runWorkspaceDelete(cmd, ns)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().Bool("dry-run", false, "show impact summary without deleting")
	cmd.Flags().Bool("cascade", false, "required for actual deletion")
	cmd.Flags().Bool("yes", false, "confirm deletion")
	return cmd
}

func runWorkspaceCreate(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runWorkspaceCreateWithStore(cmd, ns, store)
}

func runWorkspaceCreateWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return NewValidation("name is required")
	}

	ctx := context.Background()
	setup, err := rt.BootstrapService.EnsureDefaultSetup(ctx)
	if err != nil {
		return err
	}

	workspace, board, err := rt.ContextService.CreateWorkspace(ctx, setup.Provider.ID, name)
	if err != nil {
		return err
	}

	setContext, _ := cmd.Flags().GetBool("set-context")
	if setContext {
		if err := store.SetCLIContext(ns.Key, state.CLIContext{
			WorkspaceID: workspace.ID,
			BoardID:     board.ID,
		}); err != nil {
			return err
		}
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "workspace", map[string]interface{}{
			"id":       workspace.ID,
			"name":     workspace.Name,
			"board":    board.Name,
			"board_id": board.ID,
		})
	}

	return RenderWriteResult(cmd.OutOrStdout(), "workspace", workspace.ID, map[string]string{
		"Name":     workspace.Name,
		"Board":    board.Name,
		"Board ID": board.ID,
	})
}

func runWorkspaceUpdate(cmd *cobra.Command, ns Namespace) error {
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

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return NewValidation("name is required")
	}

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	ctx := context.Background()
	workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	if err != nil {
		return err
	}

	if err := rt.ContextService.RenameWorkspace(ctx, workspaceID, name); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "workspace", map[string]interface{}{
			"id":   workspaceID,
			"name": name,
		})
	}

	_, _ = cmd.OutOrStdout().Write([]byte("workspace updated\n"))
	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":   workspaceID,
		"Name": name,
	})
}

func runWorkspaceDelete(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runWorkspaceDeleteWithStore(cmd, ns, store)
}

func runWorkspaceDeleteWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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
	workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	if err != nil {
		return err
	}

	impact, err := rt.WorkspaceDeleteService.Impact(ctx, workspaceID)
	if err != nil {
		return err
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		impactMap := map[string]int{
			"boards":   impact.Boards,
			"columns":  impact.Columns,
			"tasks":    impact.Tasks,
			"comments": impact.Comments,
		}
		if cfg.JSON {
			return RenderDryRunImpactJSON(cmd.OutOrStdout(), "workspace", impactMap)
		}
		return RenderDryRunImpact(cmd.OutOrStdout(), "workspace", impactMap)
	}

	cascade, _ := cmd.Flags().GetBool("cascade")
	confirmed, _ := cmd.Flags().GetBool("yes")
	if !cascade || !confirmed {
		return NewValidation("deletion requires both --cascade and --yes")
	}

	if err := rt.WorkspaceDeleteService.Delete(ctx, workspaceID); err != nil {
		return err
	}

	if err := store.SanitizeNamespace(ns.Key, workspaceID); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderDeleteResultJSON(cmd.OutOrStdout(), "workspace", workspaceID, cascade)
	}

	return RenderDeleteResult(cmd.OutOrStdout(), "workspace", workspaceID)
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
