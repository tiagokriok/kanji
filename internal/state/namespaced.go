package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SharedState is the top-level persisted structure for all namespaces.
type SharedState struct {
	Namespaces map[string]NamespaceState `json:"namespaces"`
}

// NamespaceState holds CLI context and TUI state for a single namespace.
type NamespaceState struct {
	CLIContext CLIContext `json:"cli_context"`
	TUIState   TUIState   `json:"tui_state"`
}

// CLIContext holds the explicit CLI-selected workspace and board.
type CLIContext struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
	BoardID     string `json:"board_id,omitempty"`
}

// TUIState mirrors the existing TUI persisted state shape for compatibility.
type TUIState struct {
	LastWorkspaceID      string            `json:"last_workspace_id,omitempty"`
	LastBoardByWorkspace map[string]string `json:"last_board_by_workspace,omitempty"`
}

// Store persists SharedState as JSON per namespace.
type Store struct {
	path string
}

// NewStore creates a Store at the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// DefaultStorePath returns the canonical shared state file path.
func DefaultStorePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(cfgDir, "kanji", "namespaced_state.json"), nil
}

// Load reads the shared state file. If the file is missing, it returns an
// empty SharedState safely.
func (s *Store) Load() (SharedState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return SharedState{Namespaces: map[string]NamespaceState{}}, nil
		}
		return SharedState{}, fmt.Errorf("read state file: %w", err)
	}

	var st SharedState
	if err := json.Unmarshal(data, &st); err != nil {
		return SharedState{}, fmt.Errorf("unmarshal state: %w", err)
	}

	if st.Namespaces == nil {
		st.Namespaces = map[string]NamespaceState{}
	}

	return st, nil
}

// Save writes the shared state to disk.
func (s *Store) Save(st SharedState) error {
	if st.Namespaces == nil {
		st.Namespaces = map[string]NamespaceState{}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// GetCLIContext returns the CLI context for the given namespace.
// If the namespace does not exist, it returns an empty CLIContext.
func (s *Store) GetCLIContext(namespace string) (CLIContext, error) {
	st, err := s.Load()
	if err != nil {
		return CLIContext{}, err
	}
	ns, ok := st.Namespaces[namespace]
	if !ok {
		return CLIContext{}, nil
	}
	return ns.CLIContext, nil
}

// SetCLIContext updates the CLI context for the given namespace,
// preserving any existing TUI state.
func (s *Store) SetCLIContext(namespace string, ctx CLIContext) error {
	st, err := s.Load()
	if err != nil {
		return err
	}
	ns := st.Namespaces[namespace]
	ns.CLIContext = ctx
	st.Namespaces[namespace] = ns
	return s.Save(st)
}

// ClearCLIContext removes the CLI context for the given namespace
// while preserving its TUI state.
func (s *Store) ClearCLIContext(namespace string) error {
	st, err := s.Load()
	if err != nil {
		return err
	}
	ns := st.Namespaces[namespace]
	ns.CLIContext = CLIContext{}
	st.Namespaces[namespace] = ns
	return s.Save(st)
}

// GetTUIState returns the TUI state for the given namespace.
func (s *Store) GetTUIState(namespace string) (TUIState, error) {
	st, err := s.Load()
	if err != nil {
		return TUIState{}, err
	}
	ns, ok := st.Namespaces[namespace]
	if !ok {
		return TUIState{}, nil
	}
	return ns.TUIState, nil
}

// SetTUIState updates the TUI state for the given namespace,
// preserving any existing CLI context.
func (s *Store) SetTUIState(namespace string, ts TUIState) error {
	st, err := s.Load()
	if err != nil {
		return err
	}
	ns := st.Namespaces[namespace]
	ns.TUIState = ts
	st.Namespaces[namespace] = ns
	return s.Save(st)
}
