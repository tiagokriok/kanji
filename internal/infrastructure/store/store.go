package store

import (
	"context"

	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

// Store is a shared transaction boundary for SQLite-backed operations.
type Store interface {
	InTx(ctx context.Context, fn func(Tx) error) error
	Queries() *sqlc.Queries
}

// Tx exposes the sqlc query interface inside a transaction.
type Tx interface {
	Queries() *sqlc.Queries
}
