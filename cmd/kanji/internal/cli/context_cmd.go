package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/state"
)

func defaultStateStore() (*state.Store, error) {
	path, err := state.DefaultStorePath()
	if err != nil {
		return nil, err
	}
	return state.NewStore(path), nil
}

func newContextCommand() *cobra.Command {
	ctx := &cobra.Command{
		Use:   "context",
		Short: "Manage CLI context",
	}
	ctx.AddCommand(newContextShowCommand())
	ctx.AddCommand(newContextClearCommand())
	ctx.AddCommand(newContextSetCommand())
	return ctx
}

func newContextShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current namespace and context",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := defaultStateStore()
			if err != nil {
				return err
			}
			return runContextShow(cmd, store)
		},
	}
}

func runContextShow(cmd *cobra.Command, store *state.Store) error {
	ns, err := ResolveNamespace()
	if err != nil {
		return err
	}

	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	cliCtx, _ := store.GetCLIContext(ns.Key)
	tuiState, _ := store.GetTUIState(ns.Key)

	if cfg.JSON {
		return renderContextJSON(cmd.OutOrStdout(), ns, cliCtx, tuiState)
	}
	return renderContextHuman(cmd.OutOrStdout(), ns, cliCtx, tuiState)
}

func renderContextHuman(w io.Writer, ns Namespace, cliCtx state.CLIContext, tuiState state.TUIState) error {
	fmt.Fprintf(w, "Namespace: %s\n", ns.Key)
	fmt.Fprintf(w, "Source:    %s\n", ns.Source)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "CLI Context:")
	if cliCtx.WorkspaceID != "" {
		fmt.Fprintf(w, "  Workspace: %s\n", cliCtx.WorkspaceID)
	} else {
		fmt.Fprintln(w, "  (none)")
	}
	if cliCtx.BoardID != "" {
		fmt.Fprintf(w, "  Board:     %s\n", cliCtx.BoardID)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "TUI State:")
	if tuiState.LastWorkspaceID != "" {
		fmt.Fprintf(w, "  Last Workspace: %s\n", tuiState.LastWorkspaceID)
	} else {
		fmt.Fprintln(w, "  (none)")
	}
	return nil
}

func newContextClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear explicit CLI context for the current namespace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := defaultStateStore()
			if err != nil {
				return err
			}
			return runContextClear(cmd, store)
		},
	}
}

func runContextClear(cmd *cobra.Command, store *state.Store) error {
	ns, err := ResolveNamespace()
	if err != nil {
		return err
	}

	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	if err := store.ClearCLIContext(ns.Key); err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]string{
			"namespace": ns.Key,
			"status":    "cleared",
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "context", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cleared CLI context for namespace: %s\n", ns.Key)
	return nil
}

func newContextSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set explicit CLI context for the current namespace",
		Long: `Set the workspace and optionally the board for the current namespace.

If only a workspace is provided, any existing board context is cleared.
Board name resolution requires a workspace scope.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := ResolveConfig(cmd)
			if err != nil {
				return err
			}
			rt, err := NewRuntime(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer rt.Close()
			store, err := defaultStateStore()
			if err != nil {
				return err
			}
			return runContextSet(cmd, rt, store)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	return cmd
}

func runContextSet(cmd *cobra.Command, rt *Runtime, store *state.Store) error {
	ns, err := ResolveNamespace()
	if err != nil {
		return err
	}

	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	// Resolve workspace.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") {
		id, _ := cmd.Flags().GetString("workspace-id")
		workspaceID = id
		// Validate existence.
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
		return NewValidation("workspace-id or workspace is required")
	}

	// Resolve board.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		id, _ := cmd.Flags().GetString("board-id")
		boardID = id
		// Validate existence within workspace.
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if b.ID == boardID {
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", boardID)
		}
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
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
	}

	// Save context.
	if err := store.SetCLIContext(ns.Key, state.CLIContext{
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}); err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]string{
			"namespace":    ns.Key,
			"workspace_id": workspaceID,
		}
		if boardID != "" {
			payload["board_id"] = boardID
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "context", payload)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Context set for namespace: %s\n", ns.Key)
	fmt.Fprintf(cmd.OutOrStdout(), "Workspace: %s\n", workspaceID)
	if boardID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Board:     %s\n", boardID)
	}
	return nil
}

func renderContextJSON(w io.Writer, ns Namespace, cliCtx state.CLIContext, tuiState state.TUIState) error {
	payload := map[string]interface{}{
		"namespace": map[string]string{
			"key":    ns.Key,
			"source": ns.Source,
		},
		"cli_context": map[string]string{
			"workspace_id": cliCtx.WorkspaceID,
			"board_id":     cliCtx.BoardID,
		},
		"tui_state": map[string]interface{}{
			"last_workspace_id":       tuiState.LastWorkspaceID,
			"last_board_by_workspace": tuiState.LastBoardByWorkspace,
		},
	}
	return RenderWrappedJSON(w, "context", payload)
}
