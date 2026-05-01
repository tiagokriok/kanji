package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/tiagokriok/kanji/internal/domain"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type CreateBoardColumnInput struct {
	Name  string
	Color string
}

type ContextService struct {
	repo domain.SetupRepository
}

func NewContextService(repo domain.SetupRepository) *ContextService {
	return &ContextService{repo: repo}
}

func (s *ContextService) ListWorkspaces(ctx context.Context) ([]domain.Workspace, error) {
	return s.repo.ListWorkspaces(ctx)
}

func (s *ContextService) ListBoards(ctx context.Context, workspaceID string) ([]domain.Board, error) {
	if strings.TrimSpace(workspaceID) == "" {
		return nil, errors.New("workspace id is required")
	}
	return s.repo.ListBoards(ctx, workspaceID)
}

func (s *ContextService) ListColumns(ctx context.Context, boardID string) ([]domain.Column, error) {
	if strings.TrimSpace(boardID) == "" {
		return nil, errors.New("board id is required")
	}
	return s.repo.ListColumns(ctx, boardID)
}

func (s *ContextService) CreateWorkspace(ctx context.Context, providerID, name string) (domain.Workspace, domain.Board, error) {
	providerID = strings.TrimSpace(providerID)
	name = strings.TrimSpace(name)
	if providerID == "" {
		return domain.Workspace{}, domain.Board{}, errors.New("provider id is required")
	}
	if name == "" {
		return domain.Workspace{}, domain.Board{}, errors.New("workspace name is required")
	}

	workspace := domain.Workspace{
		ID:         uuid.NewString(),
		ProviderID: providerID,
		Name:       name,
	}
	if err := s.repo.CreateWorkspace(ctx, workspace); err != nil {
		return domain.Workspace{}, domain.Board{}, err
	}

	board, err := s.CreateBoard(ctx, workspace.ID, "Main")
	if err != nil {
		return domain.Workspace{}, domain.Board{}, err
	}
	return workspace, board, nil
}

func (s *ContextService) RenameWorkspace(ctx context.Context, workspaceID, name string) error {
	return s.repo.RenameWorkspace(ctx, workspaceID, name)
}

func (s *ContextService) CreateBoard(ctx context.Context, workspaceID, name string) (domain.Board, error) {
	defaults := defaultColumnSpecs()
	columns := make([]CreateBoardColumnInput, 0, len(defaults))
	for _, d := range defaults {
		columns = append(columns, CreateBoardColumnInput{Name: d.Name, Color: d.Color})
	}
	return s.CreateBoardWithColumns(ctx, workspaceID, name, columns)
}

func (s *ContextService) CreateBoardWithColumns(
	ctx context.Context,
	workspaceID, name string,
	columns []CreateBoardColumnInput,
) (domain.Board, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	name = strings.TrimSpace(name)
	if workspaceID == "" {
		return domain.Board{}, errors.New("workspace id is required")
	}
	if name == "" {
		return domain.Board{}, errors.New("board name is required")
	}
	if len(columns) == 0 {
		return domain.Board{}, errors.New("at least one column is required")
	}

	board := domain.Board{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		Name:        name,
		ViewDefault: "list",
	}
	if err := s.repo.CreateBoard(ctx, board); err != nil {
		return domain.Board{}, err
	}

	var createdColumns []domain.Column
	position := 1
	created := 0
	for i, input := range columns {
		columnName := strings.TrimSpace(input.Name)
		if columnName == "" {
			continue
		}
		color := strings.ToUpper(strings.TrimSpace(input.Color))
		if color == "" {
			color = NextDefaultColor(createdColumns)
		}
		if !hexColorPattern.MatchString(color) {
			return domain.Board{}, fmt.Errorf("column %d color must be HEX (#RRGGBB)", i+1)
		}

		c := domain.Column{
			ID:       uuid.NewString(),
			BoardID:  board.ID,
			Name:     columnName,
			Color:    color,
			Position: position,
		}
		if err := s.repo.CreateColumn(ctx, c); err != nil {
			return domain.Board{}, err
		}
		createdColumns = append(createdColumns, c)
		position++
		created++
	}

	if created == 0 {
		return domain.Board{}, errors.New("at least one column name is required")
	}

	return board, nil
}

func (s *ContextService) RenameBoard(ctx context.Context, boardID, name string) error {
	return s.repo.RenameBoard(ctx, boardID, name)
}

func (s *ContextService) CreateColumn(ctx context.Context, boardID, name, color string, wipLimit *int) (domain.Column, error) {
	boardID = strings.TrimSpace(boardID)
	name = strings.TrimSpace(name)
	if boardID == "" {
		return domain.Column{}, errors.New("board id is required")
	}
	if name == "" {
		return domain.Column{}, errors.New("column name is required")
	}

	columns, err := s.repo.ListColumns(ctx, boardID)
	if err != nil {
		return domain.Column{}, err
	}

	color = strings.ToUpper(strings.TrimSpace(color))
	if color == "" {
		color = NextDefaultColor(columns)
	}
	if !hexColorPattern.MatchString(color) {
		return domain.Column{}, errors.New("color must be HEX (#RRGGBB)")
	}

	position := 1
	for _, c := range columns {
		if c.Position >= position {
			position = c.Position + 1
		}
	}

	column := domain.Column{
		ID:       uuid.NewString(),
		BoardID:  boardID,
		Name:     name,
		Color:    color,
		Position: position,
		WIPLimit: wipLimit,
	}

	if err := s.repo.CreateColumn(ctx, column); err != nil {
		return domain.Column{}, err
	}

	return column, nil
}

func (s *ContextService) UpdateColumn(ctx context.Context, columnID string, name, color *string, wipLimit *int, clearWIP bool) error {
	return s.repo.UpdateColumn(ctx, columnID, name, color, wipLimit, clearWIP)
}

func (s *ContextService) ReorderColumns(ctx context.Context, boardID string, orderedColumnIDs []string) error {
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return errors.New("board id is required")
	}
	if len(orderedColumnIDs) == 0 {
		return errors.New("at least one column id is required")
	}

	existing, err := s.repo.ListColumns(ctx, boardID)
	if err != nil {
		return err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, c := range existing {
		existingSet[c.ID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(orderedColumnIDs))
	for i, id := range orderedColumnIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("column id at position %d is required", i+1)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate column id at position %d", i+1)
		}
		if _, ok := existingSet[id]; !ok {
			return fmt.Errorf("column %s not found in board", id)
		}
		seen[id] = struct{}{}
		orderedColumnIDs[i] = id
	}

	for _, c := range existing {
		if _, ok := seen[c.ID]; !ok {
			return fmt.Errorf("column %s not included in reorder", c.ID)
		}
	}

	return s.repo.ReorderColumns(ctx, boardID, orderedColumnIDs)
}

func (s *ContextService) BuildLastBoardByWorkspace(ctx context.Context) (map[string]string, error) {
	workspaces, err := s.repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(workspaces))
	for _, ws := range workspaces {
		boards, listErr := s.repo.ListBoards(ctx, ws.ID)
		if listErr != nil {
			return nil, listErr
		}
		if len(boards) > 0 {
			result[ws.ID] = boards[0].ID
		}
	}
	return result, nil
}
