package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/state"
)

// DuplicateName describes a group of entities that share the same
// normalized name within a scope (workspace, board, or column).
type DuplicateName struct {
	Scope string
	Name  string
	Count int
	IDs   []string
}

// FindDuplicateWorkspaceNames returns workspaces whose LOWER(TRIM(name))
// appears more than once.
func FindDuplicateWorkspaceNames(ctx context.Context, repo domain.SetupRepository) ([]DuplicateName, error) {
	workspaces, err := repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]string, len(workspaces))
	for _, ws := range workspaces {
		key := strings.ToLower(strings.TrimSpace(ws.Name))
		groups[key] = append(groups[key], ws.ID)
	}

	var dups []DuplicateName
	for name, ids := range groups {
		if len(ids) > 1 {
			dups = append(dups, DuplicateName{
				Scope: "workspace",
				Name:  name,
				Count: len(ids),
				IDs:   ids,
			})
		}
	}
	return dups, nil
}

// FindDuplicateBoardNames returns boards whose
// (workspace_id, LOWER(TRIM(name))) appears more than once.
func FindDuplicateBoardNames(ctx context.Context, repo domain.SetupRepository) ([]DuplicateName, error) {
	workspaces, err := repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	var dups []DuplicateName
	for _, ws := range workspaces {
		boards, err := repo.ListBoards(ctx, ws.ID)
		if err != nil {
			return nil, err
		}
		groups := make(map[string][]string, len(boards))
		for _, b := range boards {
			key := strings.ToLower(strings.TrimSpace(b.Name))
			groups[key] = append(groups[key], b.ID)
		}
		for name, ids := range groups {
			if len(ids) > 1 {
				dups = append(dups, DuplicateName{
					Scope: "board",
					Name:  name,
					Count: len(ids),
					IDs:   ids,
				})
			}
		}
	}
	return dups, nil
}

// FindDuplicateColumnNames returns columns whose
// (board_id, LOWER(TRIM(name))) appears more than once.
func FindDuplicateColumnNames(ctx context.Context, repo domain.SetupRepository) ([]DuplicateName, error) {
	workspaces, err := repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	var dups []DuplicateName
	for _, ws := range workspaces {
		boards, err := repo.ListBoards(ctx, ws.ID)
		if err != nil {
			return nil, err
		}
		for _, b := range boards {
			columns, err := repo.ListColumns(ctx, b.ID)
			if err != nil {
				return nil, err
			}
			groups := make(map[string][]string, len(columns))
			for _, c := range columns {
				key := strings.ToLower(strings.TrimSpace(c.Name))
				groups[key] = append(groups[key], c.ID)
			}
			for name, ids := range groups {
				if len(ids) > 1 {
					dups = append(dups, DuplicateName{
						Scope: "column",
						Name:  name,
						Count: len(ids),
						IDs:   ids,
					})
				}
			}
		}
	}
	return dups, nil
}

// FindDanglingContextRefs checks every namespace's CLIContext against the
// current database and returns descriptions of references that no longer exist.
func FindDanglingContextRefs(ctx context.Context, setupRepo domain.SetupRepository, stateStore *state.Store) ([]string, error) {
	st, err := stateStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	workspaces, err := setupRepo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	workspaceSet := make(map[string]struct{}, len(workspaces))
	for _, ws := range workspaces {
		workspaceSet[ws.ID] = struct{}{}
	}

	boardSet := make(map[string]struct{})
	for _, ws := range workspaces {
		boards, err := setupRepo.ListBoards(ctx, ws.ID)
		if err != nil {
			return nil, err
		}
		for _, b := range boards {
			boardSet[b.ID] = struct{}{}
		}
	}

	var dangling []string
	for nsKey, nsState := range st.Namespaces {
		if nsState.CLIContext.WorkspaceID != "" {
			if _, ok := workspaceSet[nsState.CLIContext.WorkspaceID]; !ok {
				dangling = append(dangling, fmt.Sprintf("namespace %q: workspace_id %q does not exist", nsKey, nsState.CLIContext.WorkspaceID))
			}
		}
		if nsState.CLIContext.BoardID != "" {
			if _, ok := boardSet[nsState.CLIContext.BoardID]; !ok {
				dangling = append(dangling, fmt.Sprintf("namespace %q: board_id %q does not exist", nsKey, nsState.CLIContext.BoardID))
			}
		}
	}

	return dangling, nil
}
