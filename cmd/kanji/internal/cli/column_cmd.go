package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/state"
)

func newColumnCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "column",
		Short: "Column operations",
	}
	c.AddCommand(newColumnListCommand())
	c.AddCommand(newColumnGetCommand())
	c.AddCommand(newColumnCreateCommand())
	c.AddCommand(newColumnReorderCommand())
	return c
}

func newColumnListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List columns for a board",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnList(cmd, ns)
		},
	}
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	return cmd
}

func newColumnCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new column",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnCreate(cmd, ns)
		},
	}
	cmd.Flags().String("name", "", "column name")
	cmd.Flags().String("color", "", "column color (hex)")
	cmd.Flags().Int("wip-limit", 0, "WIP limit")
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	return cmd
}

func newColumnReorderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder board columns",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnReorder(cmd, ns)
		},
	}
	cmd.Flags().StringArray("column-id", nil, "column ID (repeat for order)")
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	return cmd
}

func runColumnCreate(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runColumnCreateWithStore(cmd, ns, store)
}

func runColumnCreateWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Resolve board scope.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		boardID, _ = cmd.Flags().GetString("board-id")
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		cliCtx, _ := store.GetCLIContext(ns.Key)
		workspaceID := cliCtx.WorkspaceID
		if workspaceID == "" {
			workspaces, err := rt.ContextService.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			if len(workspaces) == 1 {
				workspaceID = workspaces[0].ID
			} else {
				return NewValidation("board-id or board with workspace scope is required")
			}
		}
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				boardID = b.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	} else {
		cliCtx, _ := store.GetCLIContext(ns.Key)
		if cliCtx.BoardID != "" {
			boardID = cliCtx.BoardID
		} else {
			return NewValidation("board scope required: use --board-id, --board, or kanji context set")
		}
	}

	if !cmd.Flags().Changed("name") {
		return NewValidation("name is required")
	}
	name, _ := cmd.Flags().GetString("name")

	var color string
	if cmd.Flags().Changed("color") {
		color, _ = cmd.Flags().GetString("color")
	}

	var wipLimit *int
	if cmd.Flags().Changed("wip-limit") {
		wip, _ := cmd.Flags().GetInt("wip-limit")
		wipLimit = &wip
	}

	column, err := rt.ContextService.CreateColumn(ctx, boardID, name, color, wipLimit)
	if err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"id":       column.ID,
			"name":     column.Name,
			"color":    column.Color,
			"position": column.Position,
		}
		if column.WIPLimit != nil {
			payload["wip_limit"] = *column.WIPLimit
		}
		return RenderWriteResultJSON(cmd.OutOrStdout(), "column", payload)
	}

	fields := map[string]string{
		"Name":     column.Name,
		"Color":    column.Color,
		"Position": fmt.Sprintf("%d", column.Position),
	}
	if column.WIPLimit != nil {
		fields["WIP Limit"] = fmt.Sprintf("%d", *column.WIPLimit)
	}
	return RenderWriteResult(cmd.OutOrStdout(), "column", column.ID, fields)
}

func runColumnReorder(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runColumnReorderWithStore(cmd, ns, store)
}

func runColumnReorderWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Resolve board scope.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		boardID, _ = cmd.Flags().GetString("board-id")
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		cliCtx, _ := store.GetCLIContext(ns.Key)
		workspaceID := cliCtx.WorkspaceID
		if workspaceID == "" {
			workspaces, err := rt.ContextService.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			if len(workspaces) == 1 {
				workspaceID = workspaces[0].ID
			} else {
				return NewValidation("board-id or board with workspace scope is required")
			}
		}
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				boardID = b.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	} else {
		cliCtx, _ := store.GetCLIContext(ns.Key)
		if cliCtx.BoardID != "" {
			boardID = cliCtx.BoardID
		} else {
			return NewValidation("board scope required: use --board-id, --board, or kanji context set")
		}
	}

	columnIDs, err := cmd.Flags().GetStringArray("column-id")
	if err != nil {
		return err
	}
	if len(columnIDs) == 0 {
		return NewValidation("at least one --column-id is required")
	}

	if err := rt.ContextService.ReorderColumns(ctx, boardID, columnIDs); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWrappedJSON(cmd.OutOrStdout(), "reorder", map[string]interface{}{
			"board_id":   boardID,
			"column_ids": columnIDs,
		})
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Columns reordered")
	return nil
}

func newColumnGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a column by ID or name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnGet(cmd, ns)
		},
	}
	cmd.Flags().String("column-id", "", "column ID")
	cmd.Flags().String("column", "", "column name")
	cmd.Flags().String("board-id", "", "board ID (required for name resolution)")
	return cmd
}

func runColumnGet(cmd *cobra.Command, ns Namespace) error {
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

	// Resolve column.
	var column domain.Column
	if cmd.Flags().Changed("column-id") {
		id, _ := cmd.Flags().GetString("column-id")
		// Search across all boards.
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
				columns, err := rt.ContextService.ListColumns(ctx, b.ID)
				if err != nil {
					return err
				}
				for _, c := range columns {
					if c.ID == id {
						column = c
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return NewNotFound("column", id)
		}
	} else if cmd.Flags().Changed("column") {
		name, _ := cmd.Flags().GetString("column")
		var boardID string
		if cmd.Flags().Changed("board-id") {
			boardID, _ = cmd.Flags().GetString("board-id")
		} else {
			return NewValidation("board-id is required for column name resolution")
		}
		columns, err := rt.ContextService.ListColumns(ctx, boardID)
		if err != nil {
			return err
		}
		found := false
		for _, c := range columns {
			if ExactMatch(c.Name, name) {
				column = c
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("column", name)
		}
	} else {
		return NewValidation("column-id or column is required")
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"id":       column.ID,
			"name":     column.Name,
			"color":    column.Color,
			"position": column.Position,
		}
		if column.WIPLimit != nil {
			payload["wip_limit"] = *column.WIPLimit
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "column", payload)
	}

	pairs := map[string]string{
		"ID":       column.ID,
		"Name":     column.Name,
		"Color":    column.Color,
		"Position": fmt.Sprintf("%d", column.Position),
	}
	if column.WIPLimit != nil {
		pairs["WIP Limit"] = fmt.Sprintf("%d", *column.WIPLimit)
	}
	return RenderKV(cmd.OutOrStdout(), pairs)
}

func runColumnList(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runColumnListWithStore(cmd, ns, store)
}

func runColumnListWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Resolve board scope.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		boardID, _ = cmd.Flags().GetString("board-id")
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		// Need workspace scope for board name resolution.
		cliCtx, _ := store.GetCLIContext(ns.Key)
		workspaceID := cliCtx.WorkspaceID
		if workspaceID == "" {
			// Try to find workspace from context or fail.
			workspaces, err := rt.ContextService.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			if len(workspaces) == 1 {
				workspaceID = workspaces[0].ID
			} else {
				return NewValidation("board-id or board with workspace scope is required")
			}
		}
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				boardID = b.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	} else {
		// Try cli_context board.
		cliCtx, _ := store.GetCLIContext(ns.Key)
		if cliCtx.BoardID != "" {
			boardID = cliCtx.BoardID
		} else {
			return NewValidation("board scope required: use --board-id, --board, or kanji context set")
		}
	}

	columns, err := rt.ContextService.ListColumns(ctx, boardID)
	if err != nil {
		return err
	}

	if cfg.JSON {
		items := make([]map[string]interface{}, len(columns))
		for i, c := range columns {
			items[i] = map[string]interface{}{
				"id":        c.ID,
				"name":      c.Name,
				"color":     c.Color,
				"position":  c.Position,
				"wip_limit": c.WIPLimit,
			}
		}
		return RenderWrappedListJSON(cmd.OutOrStdout(), "columns", items, len(columns))
	}

	headers := []string{"ID", "Name", "Color", "Position", "WIP Limit"}
	rows := make([][]string, len(columns))
	for i, c := range columns {
		wip := ""
		if c.WIPLimit != nil {
			wip = fmt.Sprintf("%d", *c.WIPLimit)
		}
		rows[i] = []string{c.ID, c.Name, c.Color, fmt.Sprintf("%d", c.Position), wip}
	}
	return RenderTable(cmd.OutOrStdout(), headers, rows)
}
