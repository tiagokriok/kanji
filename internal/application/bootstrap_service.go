package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/tiagokriok/lazytask/internal/domain"
)

type BootstrapResult struct {
	Provider  domain.Provider
	Workspace domain.Workspace
	Board     domain.Board
	Columns   []domain.Column
}

type BootstrapService struct {
	repo domain.SetupRepository
}

func NewBootstrapService(repo domain.SetupRepository) *BootstrapService {
	return &BootstrapService{repo: repo}
}

func (s *BootstrapService) EnsureDefaultSetup(ctx context.Context) (BootstrapResult, error) {
	providers, err := s.repo.ListProviders(ctx)
	if err != nil {
		return BootstrapResult{}, err
	}

	var provider domain.Provider
	if len(providers) == 0 {
		now := time.Now().UTC()
		provider = domain.Provider{
			ID:        uuid.NewString(),
			Type:      "local",
			Name:      "Local",
			CreatedAt: now,
		}
		if err := s.repo.CreateProvider(ctx, provider); err != nil {
			return BootstrapResult{}, err
		}
	} else {
		provider = providers[0]
	}

	workspaces, err := s.repo.ListWorkspaces(ctx)
	if err != nil {
		return BootstrapResult{}, err
	}
	var workspace domain.Workspace
	if len(workspaces) == 0 {
		workspace = domain.Workspace{
			ID:         uuid.NewString(),
			ProviderID: provider.ID,
			Name:       "Default Workspace",
		}
		if err := s.repo.CreateWorkspace(ctx, workspace); err != nil {
			return BootstrapResult{}, err
		}
	} else {
		workspace = workspaces[0]
	}

	boards, err := s.repo.ListBoards(ctx, workspace.ID)
	if err != nil {
		return BootstrapResult{}, err
	}
	var board domain.Board
	if len(boards) == 0 {
		board = domain.Board{
			ID:          uuid.NewString(),
			WorkspaceID: workspace.ID,
			Name:        "Default Board",
			ViewDefault: "list",
		}
		if err := s.repo.CreateBoard(ctx, board); err != nil {
			return BootstrapResult{}, err
		}
	} else {
		board = boards[0]
	}

	columns, err := s.repo.ListColumns(ctx, board.ID)
	if err != nil {
		return BootstrapResult{}, err
	}
	if len(columns) == 0 {
		defaults := []struct {
			Name     string
			Position int
		}{
			{Name: "Todo", Position: 1},
			{Name: "Doing", Position: 2},
			{Name: "Done", Position: 3},
		}
		created := make([]domain.Column, 0, len(defaults))
		for _, d := range defaults {
			c := domain.Column{
				ID:       uuid.NewString(),
				BoardID:  board.ID,
				Name:     d.Name,
				Position: d.Position,
			}
			if err := s.repo.CreateColumn(ctx, c); err != nil {
				return BootstrapResult{}, err
			}
			created = append(created, c)
		}
		columns = created
	}

	if len(columns) == 0 {
		return BootstrapResult{}, errors.New("no columns available after bootstrap")
	}

	return BootstrapResult{
		Provider:  provider,
		Workspace: workspace,
		Board:     board,
		Columns:   columns,
	}, nil
}
