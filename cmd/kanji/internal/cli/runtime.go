package cli

import (
	"context"
	"fmt"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/repositories"
	"github.com/tiagokriok/kanji/internal/infrastructure/store"
)

// Runtime holds the initialized infrastructure and application services
// for a single CLI command invocation.
type Runtime struct {
	DB               *db.SQLiteAdapter
	Store            store.Store
	BootstrapService *application.BootstrapService
	TaskService      *application.TaskService
	TaskFlow         *application.TaskFlow
	CommentService   *application.CommentService
	ContextService   *application.ContextService
}

// Close releases the database connection.
func (r *Runtime) Close() error {
	if r.DB != nil {
		return r.DB.Close()
	}
	return nil
}

// NewRuntime opens the database, runs migrations, and wires all
// application services. It does not bootstrap the system; commands
// that need bootstrap should call BootstrapService.EnsureDefaultSetup
// explicitly.
func NewRuntime(ctx context.Context, cfg RuntimeConfig) (*Runtime, error) {
	adapter, err := db.NewSQLiteAdapter(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.RunMigrations(ctx, adapter.Raw()); err != nil {
		_ = adapter.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	s := store.New(adapter)
	setupRepo := repositories.NewSetupRepository(s)
	taskRepo := repositories.NewTaskRepository(s)
	commentRepo := repositories.NewCommentRepository(s)

	rt := &Runtime{
		DB:               adapter,
		Store:            s,
		BootstrapService: application.NewBootstrapService(setupRepo),
		TaskService:      application.NewTaskService(taskRepo),
		TaskFlow:         application.NewTaskFlow(taskRepo),
		CommentService:   application.NewCommentService(commentRepo),
		ContextService:   application.NewContextService(setupRepo),
	}

	return rt, nil
}
