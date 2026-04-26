package store

import (
	"context"

	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

// Store is a shared transaction boundary for SQLite-backed operations.
type Store interface {
	InTx(ctx context.Context, fn func(Tx) error) error
}

// Tx exposes the sqlc query bound to the current transaction.
type Tx interface {
	Queries() *sqlc.Queries
}
