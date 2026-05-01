package cli

import (
	"context"
	"errors"
	"fmt"
)

// GuardBootstrap checks whether the system has been initialized.
// If not, it returns an actionable error prompting the user to run bootstrap.
func GuardBootstrap(rt *Runtime) error {
	ctx := context.Background()
	providers, err := rt.ContextService.ListWorkspaces(ctx)
	if err != nil {
		return fmt.Errorf("check bootstrap state: %w", err)
	}
	if len(providers) == 0 {
		return errors.New("kanji is not initialized. Run: kanji data bootstrap")
	}
	return nil
}
