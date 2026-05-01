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
