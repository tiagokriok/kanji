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
	c.AddCommand(newColumnUpdateCommand())
	c.AddCommand(newColumnReorderCommand())
	c.AddCommand(newColumnDeleteCommand())
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

func newColumnDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a column",
		Long: `Delete a column. If the column contains tasks, you must reassign them first
using --move-tasks-to. Requires --yes for confirmation.`,
		Example: `  # Delete empty column
  kanji column delete --column-id <id> --yes

  # Delete and move tasks
  kanji column delete --column-id <id> --move-tasks-to <other-id> --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnDelete(cmd, ns)
		},
	}
	cmd.Flags().String("column-id", "", "column ID")
	cmd.Flags().String("column", "", "column name")
	cmd.Flags().String("move-tasks-to", "", "move tasks to this column ID before deleting")
	cmd.Flags().Bool("yes", false, "confirm deletion")
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
			return NewValidation("workspace scope required: use --workspace-id, --workspace, or kanji context set")
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

func runColumnDelete(cmd *cobra.Command, ns Namespace) error {
	store, err := defaultStateStore()
	if err != nil {
		return err
	}
	return runColumnDeleteWithStore(cmd, ns, store)
}

func runColumnDeleteWithStore(cmd *cobra.Command, ns Namespace, store *state.Store) error {
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

	// Resolve board scope.
	boardID, _, err := ResolveBoardScope(cmd, rt, store, ns, workspaceID)
	if err != nil {
		return err
	}

	// Resolve column ID.
	var columnID string
	columns, err := rt.ContextService.ListColumns(ctx, boardID)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("column-id") {
		id, _ := cmd.Flags().GetString("column-id")
		found := false
		for _, c := range columns {
			if c.ID == id {
				columnID = c.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("column", id)
		}
	} else if cmd.Flags().Changed("column") {
		name, _ := cmd.Flags().GetString("column")
		var matches []domain.Column
		for _, c := range columns {
			if ExactMatch(c.Name, name) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 0 {
			return NewNotFound("column", name)
		}
		if len(matches) > 1 {
			return NewAmbiguous("column", name, len(matches))
		}
		columnID = matches[0].ID
	} else {
		return NewValidation("column-id or column is required")
	}

	count, err := rt.ColumnDeleteService.ColumnTaskCount(ctx, workspaceID, columnID)
	if err != nil {
		return err
	}

	moveTasksTo, _ := cmd.Flags().GetString("move-tasks-to")
	if count > 0 && moveTasksTo == "" {
		return NewValidation(fmt.Sprintf("column has %d tasks: use --move-tasks-to to reassign before deleting", count))
	}

	if moveTasksTo != "" {
		// Validate destination column exists in the same board.
		found := false
		for _, c := range columns {
			if c.ID == moveTasksTo {
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("column", moveTasksTo)
		}
		if err := rt.ColumnDeleteService.ReassignTasks(ctx, workspaceID, columnID, moveTasksTo); err != nil {
			return err
		}
	}

	if err := RequireConfirmation(cmd, "yes"); err != nil {
		return err
	}

	if err := rt.ColumnDeleteService.DeleteColumn(ctx, columnID); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWrappedJSON(cmd.OutOrStdout(), "column", map[string]interface{}{
			"id":     columnID,
			"status": "deleted",
		})
	}

	_, _ = cmd.OutOrStdout().Write([]byte("column deleted\n"))
	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":     columnID,
		"Status": "deleted",
	})
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
			return NewValidation("workspace scope required: use --workspace-id, --workspace, or kanji context set")
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

func newColumnUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a column by ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runColumnUpdate(cmd, ns)
		},
	}
	cmd.Flags().String("column-id", "", "column ID")
	cmd.Flags().String("column", "", "column name")
	cmd.Flags().String("name", "", "new name")
	cmd.Flags().String("color", "", "new hex color")
	cmd.Flags().Int("wip-limit", 0, "WIP limit")
	cmd.Flags().Bool("clear-wip-limit", false, "clear WIP limit")
	return cmd
}

func runColumnUpdate(cmd *cobra.Command, ns Namespace) error {
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

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	var columnID string
	if cmd.Flags().Changed("column-id") {
		columnID, _ = cmd.Flags().GetString("column-id")
	} else if cmd.Flags().Changed("column") {
		name, _ := cmd.Flags().GetString("column")
		workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
		if err != nil {
			return err
		}
		boardID, _, err := ResolveBoardScope(cmd, rt, store, ns, workspaceID)
		if err != nil {
			return err
		}
		columns, err := rt.ContextService.ListColumns(ctx, boardID)
		if err != nil {
			return err
		}
		var matches []domain.Column
		for _, c := range columns {
			if ExactMatch(c.Name, name) {
				matches = append(matches, c)
			}
		}
		if len(matches) == 0 {
			return NewNotFound("column", name)
		}
		if len(matches) > 1 {
			return NewAmbiguous("column", name, len(matches))
		}
		columnID = matches[0].ID
	} else {
		return NewValidation("column-id or column is required")
	}

	var name, color *string
	var wipLimit *int
	clearWIP := false

	if cmd.Flags().Changed("name") {
		v, _ := cmd.Flags().GetString("name")
		name = &v
	}
	if cmd.Flags().Changed("color") {
		v, _ := cmd.Flags().GetString("color")
		color = &v
	}
	if cmd.Flags().Changed("wip-limit") && cmd.Flags().Changed("clear-wip-limit") {
		return NewValidation("cannot use both --wip-limit and --clear-wip-limit")
	}
	if cmd.Flags().Changed("wip-limit") {
		v, _ := cmd.Flags().GetInt("wip-limit")
		wipLimit = &v
	}
	if cmd.Flags().Changed("clear-wip-limit") {
		clearWIP = true
	}

	if name == nil && color == nil && wipLimit == nil && !clearWIP {
		return NewValidation("at least one of --name, --color, --wip-limit, --clear-wip-limit is required")
	}

	if err := rt.ContextService.UpdateColumn(ctx, columnID, name, color, wipLimit, clearWIP); err != nil {
		return err
	}

	// Fetch updated column for output.
	var updated domain.Column
	workspaces, _ := rt.ContextService.ListWorkspaces(ctx)
	for _, ws := range workspaces {
		boards, _ := rt.ContextService.ListBoards(ctx, ws.ID)
		for _, b := range boards {
			cols, _ := rt.ContextService.ListColumns(ctx, b.ID)
			for _, c := range cols {
				if c.ID == columnID {
					updated = c
					break
				}
			}
			if updated.ID != "" {
				break
			}
		}
		if updated.ID != "" {
			break
		}
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"id":       updated.ID,
			"name":     updated.Name,
			"color":    updated.Color,
			"position": updated.Position,
		}
		if updated.WIPLimit != nil {
			payload["wip_limit"] = *updated.WIPLimit
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "column", payload)
	}

	pairs := map[string]string{
		"ID":       updated.ID,
		"Name":     updated.Name,
		"Color":    updated.Color,
		"Position": fmt.Sprintf("%d", updated.Position),
	}
	if updated.WIPLimit != nil {
		pairs["WIP Limit"] = fmt.Sprintf("%d", *updated.WIPLimit)
	}
	return RenderKV(cmd.OutOrStdout(), pairs)
}
