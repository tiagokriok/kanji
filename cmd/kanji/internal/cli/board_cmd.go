package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/state"
)

func newBoardCommand() *cobra.Command {
	b := &cobra.Command{
		Use:   "board",
		Short: "Board operations",
	}
	b.AddCommand(newBoardListCommand())
	b.AddCommand(newBoardGetCommand())
	return b
}

func newBoardListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List boards for a workspace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runBoardList(cmd, ns)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	return cmd
}

func newBoardGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a board by ID or name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runBoardGet(cmd, ns)
		},
	}
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	cmd.Flags().String("workspace-id", "", "workspace ID (required for name resolution)")
	cmd.Flags().String("workspace", "", "workspace name (required for name resolution)")
	return cmd
}

func runBoardGet(cmd *cobra.Command, ns Namespace) error {
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

	// Resolve board.
	var board domain.Board
	if cmd.Flags().Changed("board-id") {
		id, _ := cmd.Flags().GetString("board-id")
		// Need to find board across all workspaces.
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			boards, err := rt.ContextService.ListBoards(ctx, ws.ID)
			if err != nil {
				return err
			}
			for _, b := range boards {
				if b.ID == id {
					board = b
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return NewNotFound("board", id)
		}
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		// Need workspace scope.
		var workspaceID string
		if cmd.Flags().Changed("workspace-id") {
			workspaceID, _ = cmd.Flags().GetString("workspace-id")
		} else if cmd.Flags().Changed("workspace") {
			wsName, _ := cmd.Flags().GetString("workspace")
			workspaces, err := rt.ContextService.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			found := false
			for _, ws := range workspaces {
				if ExactMatch(ws.Name, wsName) {
					workspaceID = ws.ID
					found = true
					break
				}
			}
			if !found {
				return NewNotFound("workspace", wsName)
			}
		} else {
			return NewValidation("workspace scope required for board name resolution")
		}

		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				board = b
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	} else {
		return NewValidation("board-id or board is required")
	}

	if cfg.JSON {
		return RenderWrappedJSON(cmd.OutOrStdout(), "board", map[string]string{
			"id":   board.ID,
			"name": board.Name,
		})
	}

	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":   board.ID,
		"Name": board.Name,
	})
}

func runBoardList(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runBoardListWithStore(cmd, ns, store)
}

func runBoardListWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Resolve workspace scope.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") {
		id, _ := cmd.Flags().GetString("workspace-id")
		workspaceID = id
		// Validate.
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			if ws.ID == workspaceID {
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", workspaceID)
		}
	} else if cmd.Flags().Changed("workspace") {
		name, _ := cmd.Flags().GetString("workspace")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			if ExactMatch(ws.Name, name) {
				workspaceID = ws.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", name)
		}
	} else {
		// Try cli_context.
		cliCtx, _ := store.GetCLIContext(ns.Key)
		if cliCtx.WorkspaceID != "" {
			workspaceID = cliCtx.WorkspaceID
		} else {
			return NewValidation("workspace scope required: use --workspace-id, --workspace, or kanji context set")
		}
	}

	boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
	if err != nil {
		return err
	}

	if cfg.JSON {
		items := make([]map[string]string, len(boards))
		for i, b := range boards {
			items[i] = map[string]string{
				"id":   b.ID,
				"name": b.Name,
			}
		}
		return RenderWrappedListJSON(cmd.OutOrStdout(), "boards", items, len(boards))
	}

	headers := []string{"ID", "Name"}
	rows := make([][]string, len(boards))
	for i, b := range boards {
		rows[i] = []string{b.ID, b.Name}
	}
	return RenderTable(cmd.OutOrStdout(), headers, rows)
}
