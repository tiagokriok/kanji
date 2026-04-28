package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/db/sqlc"
)

func newTestAdapter(t *testing.T) db.Adapter {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "kanji-repo-test-*.db")
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

func seedProvider(t *testing.T, ctx context.Context, q *sqlc.Queries) string {
	t.Helper()
	providerID := "p-setup"
	if err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
		ID:        providerID,
		Type:      "local",
		Name:      "Test Provider",
		CreatedAt: "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("create provider: %v", err)
	}
	return providerID
}

func TestTestHelpers_SeedProvider(t *testing.T) {
	adapter := newTestAdapter(t)
	ctx := context.Background()
	q := adapter.Queries()

	id := seedProvider(t, ctx, q)
	if id != "p-setup" {
		t.Errorf("id = %q, want p-setup", id)
	}

	items, err := q.ListProviders(ctx)
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(providers) = %d, want 1", len(items))
	}
	if items[0].Name != "Test Provider" {
		t.Errorf("Name = %q, want Test Provider", items[0].Name)
	}
}
