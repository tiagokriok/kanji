package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/tiagokriok/kanji/internal/domain"
)

func (m *Model) bootstrapContexts() {
	if m.contextService == nil {
		return
	}
	_ = m.reloadContextsFromStorage()
}

func (m *Model) persistContextSelection() {
	m.state.LastWorkspaceID = m.workspaceID
	if m.state.LastBoardByWorkspace == nil {
		m.state.LastBoardByWorkspace = map[string]string{}
	}
	if m.workspaceID != "" && m.boardID != "" {
		m.state.LastBoardByWorkspace[m.workspaceID] = m.boardID
	}
	_ = savePersistedUIState(m.state)
}

func (m *Model) reloadContextsFromStorage() error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	workspaces, err := m.contextService.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	m.workspaces = workspaces
	if len(workspaces) == 0 {
		return nil
	}

	targetWorkspaceID := m.workspaceID
	if m.state.LastWorkspaceID != "" && containsWorkspace(workspaces, m.state.LastWorkspaceID) {
		targetWorkspaceID = m.state.LastWorkspaceID
	}
	if targetWorkspaceID == "" || !containsWorkspace(workspaces, targetWorkspaceID) {
		targetWorkspaceID = workspaces[0].ID
	}
	if err := m.switchWorkspace(targetWorkspaceID); err != nil {
		return err
	}
	return nil
}

func containsWorkspace(items []domain.Workspace, workspaceID string) bool {
	for _, item := range items {
		if item.ID == workspaceID {
			return true
		}
	}
	return false
}

func containsBoard(items []domain.Board, boardID string) bool {
	for _, item := range items {
		if item.ID == boardID {
			return true
		}
	}
	return false
}

func workspaceName(items []domain.Workspace, workspaceID string) string {
	for _, item := range items {
		if item.ID == workspaceID {
			return item.Name
		}
	}
	return ""
}

func boardName(items []domain.Board, boardID string) string {
	for _, item := range items {
		if item.ID == boardID {
			return item.Name
		}
	}
	return ""
}

func (m *Model) switchWorkspace(workspaceID string) error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return fmt.Errorf("workspace id is required")
	}

	if len(m.workspaces) == 0 {
		workspaces, err := m.contextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		m.workspaces = workspaces
	}
	if !containsWorkspace(m.workspaces, workspaceID) {
		return fmt.Errorf("workspace not found")
	}

	m.workspaceID = workspaceID
	m.workspaceName = workspaceName(m.workspaces, workspaceID)
	m.columnFilter = ""
	m.filterIndex = -1

	boards, err := m.contextService.ListBoards(ctx, workspaceID)
	if err != nil {
		return err
	}
	m.boards = boards
	if len(boards) == 0 {
		board, createErr := m.contextService.CreateBoard(ctx, workspaceID, "Main")
		if createErr != nil {
			return createErr
		}
		m.boards = []domain.Board{board}
	}

	targetBoardID := ""
	if saved, ok := m.state.LastBoardByWorkspace[workspaceID]; ok && containsBoard(m.boards, saved) {
		targetBoardID = saved
	}
	if targetBoardID == "" && containsBoard(m.boards, m.boardID) {
		targetBoardID = m.boardID
	}
	if targetBoardID == "" {
		targetBoardID = m.boards[0].ID
	}
	if err := m.switchBoard(targetBoardID); err != nil {
		return err
	}
	m.persistContextSelection()
	return nil
}

func (m *Model) switchBoard(boardID string) error {
	if m.contextService == nil {
		return nil
	}
	ctx := context.Background()
	boardID = strings.TrimSpace(boardID)
	if boardID == "" {
		return fmt.Errorf("board id is required")
	}

	if len(m.boards) == 0 {
		boards, err := m.contextService.ListBoards(ctx, m.workspaceID)
		if err != nil {
			return err
		}
		m.boards = boards
	}
	if !containsBoard(m.boards, boardID) {
		return fmt.Errorf("board not found")
	}

	m.boardID = boardID
	m.boardName = boardName(m.boards, boardID)
	m.columnFilter = ""
	m.filterIndex = -1

	columns, err := m.contextService.ListColumns(ctx, boardID)
	if err != nil {
		return err
	}
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Position < columns[j].Position
	})
	m.columns = columns
	m.persistContextSelection()
	return nil
}

func workspaceIndexByID(items []domain.Workspace, id string) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func boardIndexByID(items []domain.Board, id string) int {
	for i, item := range items {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func (m *Model) switchBoardByOffset(delta int) (bool, error) {
	if len(m.boards) == 0 {
		return false, nil
	}

	current := boardIndexByID(m.boards, m.boardID)
	if current < 0 {
		current = 0
	}
	next := (current + delta + len(m.boards)) % len(m.boards)
	if m.boards[next].ID == m.boardID {
		return false, nil
	}
	if err := m.switchBoard(m.boards[next].ID); err != nil {
		return false, err
	}
	return true, nil
}
