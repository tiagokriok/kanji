package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newDataCommand() *cobra.Command {
	data := &cobra.Command{
		Use:   "data",
		Short: "Data operations",
	}
	data.AddCommand(newDataBootstrapCommand())
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
