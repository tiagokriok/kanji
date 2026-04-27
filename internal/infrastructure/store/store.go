package store

import (
	"context"
	"database/sql"

	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

// Store is a shared transaction boundary for SQLite-backed operations.
type Store interface {
	InTx(ctx context.Context, fn func(Tx) error) error
	Queries() *sqlc.Queries
}

// Tx exposes the sqlc query and raw transaction for mixed usage.
type Tx interface {
	Queries() *sqlc.Queries
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
