package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	uiStateAppName  = "kanji"
	uiStateFileName = "state.json"
)

type persistedUIState struct {
	LastWorkspaceID      string            `json:"last_workspace_id"`
	LastBoardByWorkspace map[string]string `json:"last_board_by_workspace"`
}

func loadPersistedUIState() persistedUIState {
	state := persistedUIState{LastBoardByWorkspace: map[string]string{}}
	path, err := uiStatePath()
	if err != nil {
		return state
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return state
	}
	if err := json.Unmarshal(content, &state); err != nil {
		return persistedUIState{LastBoardByWorkspace: map[string]string{}}
	}
	if state.LastBoardByWorkspace == nil {
		state.LastBoardByWorkspace = map[string]string{}
	}
	return state
}

func savePersistedUIState(state persistedUIState) error {
	path, err := uiStatePath()
	if err != nil {
		return err
	}
	if state.LastBoardByWorkspace == nil {
		state.LastBoardByWorkspace = map[string]string{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create ui state dir: %w", err)
	}
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ui state: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write ui state: %w", err)
	}
	return nil
}

func uiStatePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(cfgDir, uiStateAppName, uiStateFileName), nil
}
