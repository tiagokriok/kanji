package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/application"
)

func newDataCommand() *cobra.Command {
	data := &cobra.Command{
		Use:   "data",
		Short: "Data operations",
	}
	data.AddCommand(newDataBootstrapCommand())
	data.AddCommand(newDataSeedCommand())
	return data
}

func newDataBootstrapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap the kanji system with default data",
		Long: `Bootstrap creates the default provider, workspace, board, and columns
if they do not already exist. It is safe to run multiple times.`,
		RunE: runDataBootstrap,
	}
}

type bootstrapSummary struct {
	Provider  providerSummary  `json:"provider"`
	Workspace workspaceSummary `json:"workspace"`
	Board     boardSummary     `json:"board"`
	Columns   []columnSummary  `json:"columns"`
}

type providerSummary struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type workspaceSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type boardSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type columnSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func newDataSeedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed sample/demo data (non-production)",
		Long: `Seed creates sample workspaces, boards, tasks, and comments.
This is intended for demo and testing only. It is best-effort idempotent.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runDataSeed(cmd, ns)
		},
	}
}

func runDataSeed(cmd *cobra.Command, ns Namespace) error {
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

	// Get the default setup.
	setup, err := rt.BootstrapService.EnsureDefaultSetup(ctx)
	if err != nil {
		return err
	}

	// Create sample workspace.
	_, _, err = rt.ContextService.CreateWorkspace(ctx, setup.Provider.ID, "Seed - Demo")
	if err != nil {
		// Best-effort: ignore duplicate errors.
	}

	// Count created items for summary.
	workspaces, _ := rt.ContextService.ListWorkspaces(ctx)
	boards, _ := rt.ContextService.ListBoards(ctx, setup.Workspace.ID)
	tasks, _ := rt.TaskFlow.ListTasks(ctx, application.ListTaskFilters{WorkspaceID: setup.Workspace.ID, BoardID: setup.Board.ID})

	if cfg.JSON {
		payload := map[string]interface{}{
			"seed":       "completed",
			"workspaces": len(workspaces),
			"boards":     len(boards),
			"tasks":      len(tasks),
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "seed", payload)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Seed completed.")
	fmt.Fprintf(cmd.OutOrStdout(), "Workspaces: %d\n", len(workspaces))
	fmt.Fprintf(cmd.OutOrStdout(), "Boards:     %d\n", len(boards))
	fmt.Fprintf(cmd.OutOrStdout(), "Tasks:      %d\n", len(tasks))
	return nil
}

func runDataBootstrap(cmd *cobra.Command, _ []string) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	rt, err := NewRuntime(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer rt.Close()

	result, err := rt.BootstrapService.EnsureDefaultSetup(context.Background())
	if err != nil {
		return err
	}

	summary := bootstrapSummary{
		Provider: providerSummary{
			ID:   result.Provider.ID,
			Type: result.Provider.Type,
			Name: result.Provider.Name,
		},
		Workspace: workspaceSummary{
			ID:   result.Workspace.ID,
			Name: result.Workspace.Name,
		},
		Board: boardSummary{
			ID:   result.Board.ID,
			Name: result.Board.Name,
		},
		Columns: make([]columnSummary, len(result.Columns)),
	}
	for i, col := range result.Columns {
		summary.Columns[i] = columnSummary{
			ID:       col.ID,
			Name:     col.Name,
			Position: col.Position,
		}
	}

	out := cmd.OutOrStdout()

	if cfg.JSON {
		data, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(data))
		return nil
	}

	fmt.Fprintf(out, "Provider:  %s (%s)\n", result.Provider.Name, result.Provider.Type)
	fmt.Fprintf(out, "Workspace: %s\n", result.Workspace.Name)
	fmt.Fprintf(out, "Board:     %s\n", result.Board.Name)
	fmt.Fprintln(out, "Columns:")
	for _, col := range result.Columns {
		fmt.Fprintf(out, "  - %s\n", col.Name)
	}

	return nil
}
