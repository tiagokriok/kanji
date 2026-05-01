package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/application"
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
	b.AddCommand(newBoardCreateCommand())
	b.AddCommand(newBoardUpdateCommand())
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

func newBoardCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new board",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runBoardCreate(cmd, ns)
		},
	}
	cmd.Flags().String("name", "", "board name (required)")
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().StringArray("column", nil, `column spec "Name:#RRGGBB" (can be repeated)`)
	cmd.Flags().Bool("set-context", false, "set board context after creation")
	return cmd
}

func newBoardUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a board name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runBoardUpdate(cmd, ns)
		},
	}
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	cmd.Flags().String("workspace-id", "", "workspace ID (required for name resolution)")
	cmd.Flags().String("workspace", "", "workspace name (required for name resolution)")
	cmd.Flags().String("name", "", "new board name (required)")
	return cmd
}

func runBoardCreate(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runBoardCreateWithStore(cmd, ns, store)
}

func runBoardCreateWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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
	workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	if err != nil {
		return err
	}

	// Validate name.
	if !cmd.Flags().Changed("name") {
		return NewValidation("name is required")
	}
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return NewValidation("name is required")
	}

	// Parse columns.
	var board domain.Board
	columnRaws, _ := cmd.Flags().GetStringArray("column")
	if len(columnRaws) > 0 {
		specs, err := ParseColumnSpecs(columnRaws)
		if err != nil {
			return err
		}
		inputs := make([]application.CreateBoardColumnInput, len(specs))
		for i, s := range specs {
			inputs[i] = application.CreateBoardColumnInput{Name: s.Name, Color: s.Color}
		}
		board, err = rt.ContextService.CreateBoardWithColumns(ctx, workspaceID, name, inputs)
		if err != nil {
			return err
		}
	} else {
		board, err = rt.ContextService.CreateBoard(ctx, workspaceID, name)
		if err != nil {
			return err
		}
	}

	// Optionally set context.
	setCtx, _ := cmd.Flags().GetBool("set-context")
	if setCtx {
		if err := store.SetCLIContext(ns.Key, state.CLIContext{
			WorkspaceID: workspaceID,
			BoardID:     board.ID,
		}); err != nil {
			return err
		}
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "board", map[string]interface{}{
			"id":   board.ID,
			"name": board.Name,
		})
	}

	return RenderWriteResult(cmd.OutOrStdout(), "Board", board.ID, map[string]string{
		"Name": board.Name,
	})
}

func runBoardUpdate(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runBoardUpdateWithStore(cmd, ns, store)
}

func runBoardUpdateWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Validate name.
	if !cmd.Flags().Changed("name") {
		return NewValidation("name is required")
	}
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return NewValidation("name is required")
	}

	// Resolve workspace scope.
	workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	if err != nil {
		return err
	}

	// Resolve board scope.
	boardID, _, err := ResolveBoardScope(cmd, rt, store, ns, workspaceID)
	if err != nil {
		return err
	}

	if err := rt.ContextService.RenameBoard(ctx, boardID, name); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "board", map[string]interface{}{
			"id":   boardID,
			"name": name,
		})
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Board updated")
	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":   boardID,
		"Name": name,
	})
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
