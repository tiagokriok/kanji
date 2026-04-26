package store

import (
	"context"
	"fmt"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

type kernel struct {
	adapter db.Adapter
}

// New creates a Store backed by the given adapter.
func New(adapter db.Adapter) Store {
	return &kernel{adapter: adapter}
}

func (k *kernel) InTx(ctx context.Context, fn func(Tx) error) error {
	tx, err := k.adapter.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := k.adapter.Queries().WithTx(tx)
	bound := &txBound{queries: qtx}

	if err := fn(bound); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type txBound struct {
	queries *sqlc.Queries
}

func (t *txBound) Queries() *sqlc.Queries {
	return t.queries
}
