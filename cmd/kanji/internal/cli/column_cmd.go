package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newColumnCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "column",
		Short: "Column operations",
	}
	c.AddCommand(newColumnListCommand())
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

func runColumnList(cmd *cobra.Command, ns Namespace) error {
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
		store, _ := defaultStateStore()
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
		store, _ := defaultStateStore()
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
