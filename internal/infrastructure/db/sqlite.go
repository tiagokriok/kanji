package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/tiagokriok/lazytask/internal/infrastructure/db/sqlc"
)

const DefaultAppName = "lazytask"

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Adapter interface {
	Queries() *sqlc.Queries
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Raw() *sql.DB
	Close() error
}

type SQLiteAdapter struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSQLiteAdapter(dbPath string) (*SQLiteAdapter, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &SQLiteAdapter{
		db:      db,
		queries: sqlc.New(db),
	}, nil
}

func (a *SQLiteAdapter) Queries() *sqlc.Queries {
	return a.queries
}

func (a *SQLiteAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return a.db.BeginTx(ctx, opts)
}

func (a *SQLiteAdapter) Raw() *sql.DB {
	return a.db
}

func (a *SQLiteAdapter) Close() error {
	return a.db.Close()
}

func RunMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("run goose migrations: %w", err)
	}
	return nil
}

func DefaultDBPath(appName string) (string, error) {
	if appName == "" {
		appName = DefaultAppName
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(cfgDir, appName, "app.db"), nil
}
