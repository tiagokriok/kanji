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

func TestKernel_Queries_ReturnsWorkingQueries(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	if err := s.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        "p-queries",
		Type:      "local",
		Name:      "Query Test",
		CreatedAt: "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create provider via queries: %v", err)
	}

	items, err := s.Queries().ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers via queries: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(items))
	}
	if items[0].Name != "Query Test" {
		t.Errorf("Name = %q, want %q", items[0].Name, "Query Test")
	}
}

func TestKernel_InTx_BeginTxError(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.InTx(ctx, func(tx Tx) error {
		return tx.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        "p-begin-err",
			Type:      "local",
			Name:      "Should Not Exist",
			CreatedAt: "2024-01-01T00:00:00Z",
		})
	})
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}

	items, err := adapter.Queries().ListProviders(context.Background())
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if _, ok := findProvider(items, "p-begin-err"); ok {
		t.Fatal("expected provider p-begin-err to not exist")
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

func TestKernel_Write_CommitsOnSuccess(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	err := s.Write(ctx, "create provider", func(tx Tx) error {
		return tx.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        "p-write",
			Type:      "local",
			Name:      "Write Provider",
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
	p, ok := findProvider(items, "p-write")
	if !ok {
		t.Fatal("expected provider p-write to exist after commit")
	}
	if p.Name != "Write Provider" {
		t.Errorf("Name = %q, want %q", p.Name, "Write Provider")
	}
}

func TestKernel_Write_WrapsError(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	innerErr := errors.New("db down")
	err := s.Write(ctx, "create provider", func(tx Tx) error {
		return innerErr
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "create provider: db down"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
	if !errors.Is(err, innerErr) {
		t.Errorf("expected wrapped error to contain innerErr")
	}
}

func TestKernel_Write_RollbacksOnError(t *testing.T) {
	adapter := newTestAdapter(t)
	s := New(adapter)

	ctx := context.Background()
	expectedErr := errors.New("intentional failure")
	err := s.Write(ctx, "seed", func(tx Tx) error {
		if err := tx.Queries().CreateProvider(ctx, sqlc.CreateProviderParams{
			ID:        "p-write-rollback",
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
	if _, ok := findProvider(items, "p-write-rollback"); ok {
		t.Fatal("expected provider p-write-rollback to be rolled back")
	}
}
