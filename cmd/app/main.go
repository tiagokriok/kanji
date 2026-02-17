package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/lazytask/internal/application"
	"github.com/tiagokriok/lazytask/internal/infrastructure/db"
	"github.com/tiagokriok/lazytask/internal/infrastructure/providers"
	"github.com/tiagokriok/lazytask/internal/infrastructure/repositories"
	"github.com/tiagokriok/lazytask/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	defaultPath, err := db.DefaultDBPath(db.DefaultAppName)
	if err != nil {
		return err
	}

	dbPath := flag.String("db-path", defaultPath, "path to SQLite database")
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	seedOnly := flag.Bool("seed", false, "seed default/sample data and exit")
	flag.Parse()

	adapter, err := db.NewSQLiteAdapter(*dbPath)
	if err != nil {
		return err
	}
	defer adapter.Close()

	ctx := context.Background()
	if err := db.RunMigrations(ctx, adapter.Raw()); err != nil {
		return err
	}
	if *migrateOnly && !*seedOnly {
		fmt.Println("migrations completed")
		return nil
	}

	setupRepo := repositories.NewSetupRepository(adapter)
	bootstrapService := application.NewBootstrapService(setupRepo)
	setup, err := bootstrapService.EnsureDefaultSetup(ctx)
	if err != nil {
		return err
	}

	localProvider := providers.NewLocalProvider()
	if localProvider.Type() != setup.Provider.Type {
		return fmt.Errorf("provider mismatch: expected %s got %s", localProvider.Type(), setup.Provider.Type)
	}

	taskRepo := repositories.NewTaskRepository(adapter)
	commentRepo := repositories.NewCommentRepository(adapter)
	taskService := application.NewTaskService(taskRepo)
	commentService := application.NewCommentService(commentRepo)

	if *seedOnly {
		if err := seedSampleData(ctx, taskService, setup); err != nil {
			return err
		}
		fmt.Println("seed completed")
		return nil
	}

	model := ui.NewModel(taskService, commentService, setup)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func seedSampleData(ctx context.Context, taskService *application.TaskService, setup application.BootstrapResult) error {
	existing, err := taskService.ListTasks(ctx, application.ListTaskFilters{WorkspaceID: setup.Workspace.ID})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	now := time.Now().UTC()
	types := []struct {
		Title       string
		Description string
		Priority    int
		DueOffset   int
		ColumnIdx   int
	}{
		{Title: "Capture quick wins", Description: "- Add your first tasks\n- Keep work local", Priority: 1, DueOffset: 1, ColumnIdx: 0},
		{Title: "Ship MVP", Description: "## Goal\nDeliver list + kanban + details.", Priority: 2, DueOffset: 3, ColumnIdx: 1},
		{Title: "Write release notes", Description: "Document key bindings and usage.", Priority: 0, DueOffset: 7, ColumnIdx: 2},
	}

	for _, t := range types {
		var colID *string
		var status *string
		if t.ColumnIdx < len(setup.Columns) {
			cid := setup.Columns[t.ColumnIdx].ID
			st := setup.Columns[t.ColumnIdx].Name
			colID = &cid
			status = &st
		}
		due := now.AddDate(0, 0, t.DueOffset)
		boardID := setup.Board.ID
		_, err := taskService.CreateTask(ctx, application.CreateTaskInput{
			ProviderID:    setup.Provider.ID,
			WorkspaceID:   setup.Workspace.ID,
			BoardID:       &boardID,
			ColumnID:      colID,
			Title:         t.Title,
			DescriptionMD: t.Description,
			Priority:      t.Priority,
			DueAt:         &due,
			Status:        status,
			Labels:        []string{"seed"},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
