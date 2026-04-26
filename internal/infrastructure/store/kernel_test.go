package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

func newTestAdapter(t *testing.T) db.Adapter {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "kanji-store-test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })
	tmpFile.Close()

	adapter, err := db.NewSQLiteAdapter(tmpFile.Name())
	if err != nil {
		t.Fatalf("new sqlite adapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })

	if err := db.RunMigrations(context.Background(), adapter.Raw()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return adapter
}

func findProvider(items []sqlc.Provider, id string) (sqlc.Provider, bool) {
	for _, p := range items {
		if p.ID == id {
			return p, true
		}
	}
	return sqlc.Provider{}, false
}

func TestKernel_InTx_CommitsOnSuccess(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	err := s.InTx(ctx, func(tx Tx) error {
		return tx.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        "p1",
			Type:      "local",
			Name:      "Test Provider",
			CreatedAt: "2024-01-01T00:00:00Z",
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, err := adapter.Queries().ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	p, ok := findProvider(items, "p1")
	if !ok {
		t.Fatal("expected provider p1 to exist after commit")
	}
	if p.Name != "Test Provider" {
		t.Errorf("Name = %q, want %q", p.Name, "Test Provider")
	}
}

func TestKernel_InTx_RollbacksOnError(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	expectedErr := errors.New("intentional failure")
	err := s.InTx(ctx, func(tx Tx) error {
		if err := tx.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        "p2",
			Type:      "local",
			Name:      "Should Not Exist",
			CreatedAt: "2024-01-01T00:00:00Z",
		}); err != nil {
			return err
		}
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	items, err := adapter.Queries().ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if _, ok := findProvider(items, "p2"); ok {
		t.Fatal("expected provider p2 to be rolled back")
	}
}
