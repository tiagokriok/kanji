package store

import (
	"context"
	"database/sql"
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

func (k *kernel) Queries() *sqlc.Queries {
	return k.adapter.Queries()
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
	bound := &txBound{queries: qtx, tx: tx}

	if err := fn(bound); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (k *kernel) Write(ctx context.Context, op string, fn func(Tx) error) error {
	if err := k.InTx(ctx, fn); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

type txBound struct {
	queries *sqlc.Queries
	tx      *sql.Tx
}

func (t *txBound) Queries() *sqlc.Queries {
	return t.queries
}
