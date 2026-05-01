package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/state"
)

// ResolveWorkspaceScope resolves the workspace ID for write commands using
// the following precedence:
//
//  1. --workspace-id flag
//  2. --workspace name flag (exact normalized match)
//  3. explicit cli_context workspace (from state store)
//  4. fail with actionable error
//
// It returns the resolved ID and a source string ("flag" or "context") for
// metadata. TUI state is never consulted.
func ResolveWorkspaceScope(cmd *cobra.Command, rt *Runtime, store *state.Store, ns Namespace) (string, string, error) {
	ctx := context.Background()

	if cmd.Flags().Changed("workspace-id") {
		id, _ := cmd.Flags().GetString("workspace-id")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return "", "", fmt.Errorf("list workspaces: %w", err)
		}
		for _, ws := range workspaces {
			if ws.ID == id {
				return id, "flag", nil
			}
		}
		return "", "", NewNotFound("workspace", id)
	}

	if cmd.Flags().Changed("workspace") {
		name, _ := cmd.Flags().GetString("workspace")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return "", "", fmt.Errorf("list workspaces: %w", err)
		}
		for _, ws := range workspaces {
			if ExactMatch(ws.Name, name) {
				return ws.ID, "flag", nil
			}
		}
		return "", "", NewNotFound("workspace", name)
	}

	cliCtx, err := store.GetCLIContext(ns.Key)
	if err != nil {
		return "", "", fmt.Errorf("load cli context: %w", err)
	}
	if cliCtx.WorkspaceID != "" {
		workspaces, err := rt.ContextService.ListWorkspaces(context.Background())
		if err != nil {
			return "", "", fmt.Errorf("list workspaces: %w", err)
		}
		for _, ws := range workspaces {
			if ws.ID == cliCtx.WorkspaceID {
				return cliCtx.WorkspaceID, "context", nil
			}
		}
		return "", "", NewValidation("stored workspace no longer exists; run 'kanji context set'")
	}

	return "", "", NewValidation("workspace is required: provide --workspace-id, --workspace, or set context with 'kanji context set'")
}

// ResolveBoardScope resolves the board ID for write commands using the
// following precedence:
//
//  1. --board-id flag
//  2. --board name flag (exact normalized match within workspace)
//  3. explicit cli_context board (from state store, only if workspace matches)
//  4. fail with actionable error
//
// It returns the resolved ID and a source string ("flag" or "context") for
// metadata. TUI state is never consulted.
func ResolveBoardScope(cmd *cobra.Command, rt *Runtime, store *state.Store, ns Namespace, workspaceID string) (string, string, error) {
	ctx := context.Background()

	if cmd.Flags().Changed("board-id") {
		id, _ := cmd.Flags().GetString("board-id")
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return "", "", fmt.Errorf("list boards: %w", err)
		}
		for _, b := range boards {
			if b.ID == id {
				return id, "flag", nil
			}
		}
		return "", "", NewNotFound("board", id)
	}

	if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return "", "", fmt.Errorf("list boards: %w", err)
		}
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				return b.ID, "flag", nil
			}
		}
		return "", "", NewNotFound("board", name)
	}

	cliCtx, err := store.GetCLIContext(ns.Key)
	if err != nil {
		return "", "", fmt.Errorf("load cli context: %w", err)
	}
	if cliCtx.WorkspaceID == workspaceID && cliCtx.BoardID != "" {
		boards, err := rt.ContextService.ListBoards(context.Background(), workspaceID)
		if err != nil {
			return "", "", fmt.Errorf("list boards: %w", err)
		}
		for _, b := range boards {
			if b.ID == cliCtx.BoardID {
				return cliCtx.BoardID, "context", nil
			}
		}
		return "", "", NewValidation("stored board no longer exists; run 'kanji context set'")
	}

	return "", "", NewValidation("board is required: provide --board-id, --board, or set context with 'kanji context set'")
}
