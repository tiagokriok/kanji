package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
	"github.com/tiagokriok/kanji/internal/infrastructure/db"
	"github.com/tiagokriok/kanji/internal/infrastructure/providers"
	"github.com/tiagokriok/kanji/internal/infrastructure/repositories"
	"github.com/tiagokriok/kanji/internal/ui"
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
	contextService := application.NewContextService(setupRepo)

	if *seedOnly {
		if err := seedSampleData(ctx, taskService, commentService, contextService, setup); err != nil {
			return err
		}
		fmt.Println("seed completed")
		return nil
	}

	model := ui.NewModel(taskService, commentService, contextService, setup)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

type seedComment struct {
	Author string
	Body   string
}

type seedTask struct {
	Title         string
	DescriptionMD string
	Column        string
	Priority      int
	DueOffsetDays *int
	Labels        []string
	Comments      []seedComment
}

type seedBoard struct {
	Name  string
	Tasks []seedTask
}

type seedWorkspace struct {
	Name   string
	Boards []seedBoard
}

func seedSampleData(
	ctx context.Context,
	taskService *application.TaskService,
	commentService *application.CommentService,
	contextService *application.ContextService,
	setup application.BootstrapResult,
) error {
	now := time.Now().UTC()
	spec := []seedWorkspace{
		{
			Name: setup.Workspace.Name,
			Boards: []seedBoard{
				{
					Name: setup.Board.Name,
					Tasks: []seedTask{
						{
							Title:         "Critical: validate rollback path in production",
							DescriptionMD: "## Objective\nEnsure rollback is documented and tested.\n\n- Validate DB backup restore\n- Validate feature flag rollback\n- Capture runbook updates",
							Column:        "Doing",
							Priority:      0,
							DueOffsetDays: intPtr(-1),
							Labels:        []string{"release", "ops"},
							Comments: []seedComment{
								{Author: "oncall", Body: "Rollback drill started. Capturing timings and side effects."},
								{Author: "sre", Body: "Need to verify queue drain before rollback."},
							},
						},
						{
							Title:         "Markdown showcase: full renderer coverage",
							DescriptionMD: seedMarkdownShowcaseDescription(),
							Column:        "Todo",
							Priority:      2,
							DueOffsetDays: intPtr(4),
							Labels:        []string{"markdown", "demo"},
							Comments: []seedComment{
								{
									Author: "maintainer",
									Body:   seedMarkdownShowcaseCommentOne(),
								},
								{
									Author: "qa",
									Body:   seedMarkdownShowcaseCommentTwo(),
								},
							},
						},
						{
							Title:         "Urgent: stabilize flaky integration tests",
							DescriptionMD: "Failing suite blocks merges.\n\n```bash\ngo test ./... -run Integration\n```",
							Column:        "Todo",
							Priority:      1,
							DueOffsetDays: intPtr(1),
							Labels:        []string{"quality", "ci"},
						},
						{
							Title:         "High: publish release notes for v0.2",
							DescriptionMD: "### Notes\nSummarize UX changes, keybindings, and known limitations.",
							Column:        "Done",
							Priority:      2,
							DueOffsetDays: intPtr(3),
							Labels:        []string{"docs"},
							Comments: []seedComment{
								{Author: "pm", Body: "Include screenshots for list and filter panel."},
							},
						},
						{
							Title:         "Medium: document workspace and board strategy",
							DescriptionMD: "Create short guide for teams:\n- when to create workspace\n- when to create board",
							Column:        "Todo",
							Priority:      3,
							DueOffsetDays: intPtr(7),
							Labels:        []string{"docs", "ux"},
						},
						{
							Title:         "Low: clean up deprecated migration notes",
							DescriptionMD: "Remove stale notes after verifying historical references.",
							Column:        "Done",
							Priority:      4,
							DueOffsetDays: nil,
							Labels:        []string{"cleanup"},
						},
						{
							Title:         "None: backlog idea - keyboard macro mode",
							DescriptionMD: "Potential future enhancement for power users.",
							Column:        "Todo",
							Priority:      5,
							DueOffsetDays: intPtr(30),
							Labels:        []string{"idea"},
						},
					},
				},
			},
		},
		{
			Name: "Seed - Product",
			Boards: []seedBoard{
				{
					Name: "Roadmap Q2",
					Tasks: []seedTask{
						{
							Title:         "Launch context switcher UX",
							DescriptionMD: "### Success Criteria\n- Workspace switch in <2 keypresses\n- Board switch in <2 keypresses\n- Persist last context",
							Column:        "Doing",
							Priority:      1,
							DueOffsetDays: intPtr(2),
							Labels:        []string{"ux", "navigation"},
							Comments: []seedComment{
								{Author: "design", Body: "Need clear focus state in modal list."},
							},
						},
						{
							Title:         "Implement workspace permissions model (future)",
							DescriptionMD: "Exploratory task. Not for MVP.\n\n- Define roles\n- Define ownership boundaries",
							Column:        "Todo",
							Priority:      3,
							DueOffsetDays: nil,
							Labels:        []string{"architecture"},
						},
						{
							Title:         "Finalize roadmap review deck",
							DescriptionMD: "Prepare meeting-ready deck with milestones and risks.",
							Column:        "Done",
							Priority:      2,
							DueOffsetDays: intPtr(5),
							Labels:        []string{"planning"},
						},
					},
				},
				{
					Name: "Bug Triage",
					Tasks: []seedTask{
						{
							Title:         "Fix filter panel not refreshing list",
							DescriptionMD: "Regression appears when modal intercepts update messages.",
							Column:        "Done",
							Priority:      0,
							DueOffsetDays: intPtr(-2),
							Labels:        []string{"bug", "ui"},
						},
						{
							Title:         "Investigate table row wrapping in split view",
							DescriptionMD: "Reproduce on narrow terminals and verify pri column alignment.",
							Column:        "Doing",
							Priority:      1,
							DueOffsetDays: intPtr(1),
							Labels:        []string{"bug", "layout"},
						},
						{
							Title:         "Re-test markdown editor multiline persistence",
							DescriptionMD: "Validate Ctrl+g flow with 30+ lines of markdown.",
							Column:        "Todo",
							Priority:      2,
							DueOffsetDays: intPtr(4),
							Labels:        []string{"bug", "editor"},
						},
					},
				},
			},
		},
		{
			Name: "Seed - Personal",
			Boards: []seedBoard{
				{
					Name: "Home Ops",
					Tasks: []seedTask{
						{
							Title:         "Pay utilities and reconcile receipts",
							DescriptionMD: "Track water, power, and internet payment confirmation IDs.",
							Column:        "Todo",
							Priority:      2,
							DueOffsetDays: intPtr(2),
							Labels:        []string{"home", "finance"},
						},
						{
							Title:         "Schedule annual health check",
							DescriptionMD: "Call clinic and confirm available dates.",
							Column:        "Doing",
							Priority:      3,
							DueOffsetDays: intPtr(10),
							Labels:        []string{"health"},
						},
						{
							Title:         "Archive old paper documents",
							DescriptionMD: "Scan and tag all docs older than 2 years.",
							Column:        "Done",
							Priority:      5,
							DueOffsetDays: nil,
							Labels:        []string{"home", "cleanup"},
						},
					},
				},
				{
					Name: "Learning Lab",
					Tasks: []seedTask{
						{
							Title:         "Read Bubble Tea internals article",
							DescriptionMD: "Focus on update loop and command scheduling model.",
							Column:        "Todo",
							Priority:      4,
							DueOffsetDays: nil,
							Labels:        []string{"learning", "go"},
						},
						{
							Title:         "Practice SQL indexing patterns",
							DescriptionMD: "Test explain plans for workspace+board filters.",
							Column:        "Doing",
							Priority:      2,
							DueOffsetDays: intPtr(6),
							Labels:        []string{"learning", "sql"},
						},
					},
				},
			},
		},
	}

	for _, wsSpec := range spec {
		workspace, err := ensureWorkspaceByName(ctx, contextService, setup.Provider.ID, wsSpec.Name)
		if err != nil {
			return err
		}

		for _, boardSpec := range wsSpec.Boards {
			board, err := ensureBoardByName(ctx, contextService, workspace.ID, boardSpec.Name)
			if err != nil {
				return err
			}

			columns, err := contextService.ListColumns(ctx, board.ID)
			if err != nil {
				return err
			}
			existingTasks, err := taskService.ListTasks(ctx, application.ListTaskFilters{
				WorkspaceID: workspace.ID,
				BoardID:     board.ID,
			})
			if err != nil {
				return err
			}

			taskByTitle := make(map[string]domain.Task, len(existingTasks))
			for _, task := range existingTasks {
				taskByTitle[strings.ToLower(strings.TrimSpace(task.Title))] = task
			}

			for _, taskSpec := range boardSpec.Tasks {
				key := strings.ToLower(strings.TrimSpace(taskSpec.Title))
				task, exists := taskByTitle[key]
				if !exists {
					var dueAt *time.Time
					if taskSpec.DueOffsetDays != nil {
						d := now.AddDate(0, 0, *taskSpec.DueOffsetDays)
						dueAt = &d
					}

					columnID, status := resolveSeedColumn(columns, taskSpec.Column)
					task, err = taskService.CreateTask(ctx, application.CreateTaskInput{
						ProviderID:    setup.Provider.ID,
						WorkspaceID:   workspace.ID,
						BoardID:       &board.ID,
						ColumnID:      columnID,
						Title:         taskSpec.Title,
						DescriptionMD: taskSpec.DescriptionMD,
						Status:        status,
						Priority:      taskSpec.Priority,
						DueAt:         dueAt,
						Labels:        taskSpec.Labels,
					})
					if err != nil {
						return err
					}
				}

				if len(taskSpec.Comments) == 0 {
					continue
				}
				existingComments, err := commentService.ListComments(ctx, task.ID)
				if err != nil {
					return err
				}
				commentSet := make(map[string]struct{}, len(existingComments))
				for _, c := range existingComments {
					commentSet[strings.TrimSpace(c.BodyMD)] = struct{}{}
				}
				for _, c := range taskSpec.Comments {
					body := strings.TrimSpace(c.Body)
					if body == "" {
						continue
					}
					if _, ok := commentSet[body]; ok {
						continue
					}
					author := strPtr(c.Author)
					if strings.TrimSpace(c.Author) == "" {
						author = nil
					}
					if _, err := commentService.AddComment(ctx, application.AddCommentInput{
						TaskID:     task.ID,
						ProviderID: setup.Provider.ID,
						BodyMD:     body,
						Author:     author,
					}); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func ensureWorkspaceByName(
	ctx context.Context,
	contextService *application.ContextService,
	providerID string,
	name string,
) (domain.Workspace, error) {
	workspaces, err := contextService.ListWorkspaces(ctx)
	if err != nil {
		return domain.Workspace{}, err
	}
	for _, workspace := range workspaces {
		if strings.EqualFold(strings.TrimSpace(workspace.Name), strings.TrimSpace(name)) {
			return workspace, nil
		}
	}
	workspace, _, err := contextService.CreateWorkspace(ctx, providerID, name)
	return workspace, err
}

func ensureBoardByName(
	ctx context.Context,
	contextService *application.ContextService,
	workspaceID string,
	name string,
) (domain.Board, error) {
	boards, err := contextService.ListBoards(ctx, workspaceID)
	if err != nil {
		return domain.Board{}, err
	}
	for _, board := range boards {
		if strings.EqualFold(strings.TrimSpace(board.Name), strings.TrimSpace(name)) {
			return board, nil
		}
	}
	return contextService.CreateBoard(ctx, workspaceID, name)
}

func resolveSeedColumn(columns []domain.Column, preferred string) (*string, *string) {
	for _, col := range columns {
		if strings.EqualFold(strings.TrimSpace(col.Name), strings.TrimSpace(preferred)) {
			id := col.ID
			status := strings.ToLower(col.Name)
			return &id, &status
		}
	}
	if len(columns) == 0 {
		return nil, nil
	}
	id := columns[0].ID
	status := strings.ToLower(columns[0].Name)
	return &id, &status
}

func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}

func seedMarkdownShowcaseDescription() string {
	return `# Markdown Showcase

This seeded task validates the markdown renderer in the task viewer.

## Headings

### H3 Title
#### H4 Title
##### H5 Title
###### H6 Title

## Text Styling

Regular text with **bold**, *italic*, ***bold italic***, ~~strikethrough~~, and inline code like ` + "`SELECT 1`" + `.

## Lists

- Unordered item one
- Unordered item two
  - Nested child A
  - Nested child B

1. Ordered item one
2. Ordered item two
3. Ordered item three

- [ ] Task list unchecked
- [x] Task list checked

## Blockquote

> "Small steps every day create large outcomes over time."
>
> Keep the workflow simple and consistent.

## Horizontal Rule

---

## Code Blocks

~~~go
package main

import "fmt"

func main() {
    fmt.Println("kanji markdown demo")
}
~~~

~~~sql
SELECT id, title, priority
FROM tasks
WHERE priority <= 2
ORDER BY updated_at DESC;
~~~

## Table

| Field | Example | Notes |
| --- | --- | --- |
| Status | Doing | Uses board columns |
| Priority | 0-5 | Lower is more urgent |
| Due date | 2026-02-21 | Localized in UI |

## Link + Image Syntax

- Docs: [Charmbracelet](https://github.com/charmbracelet)
- Image syntax sample: ![Kanji Logo](https://example.com/logo.png)

## Mixed Content

Use this section to verify wrapping behavior on narrow terminals. This sentence is intentionally long so line breaks, spacing, and clipping behavior can be validated across multiple terminal widths without manual editing.
`
}

func seedMarkdownShowcaseCommentOne() string {
	return `### Review Notes

Great progress on the task viewer.

- Verified heading rendering
- Verified list wrapping
- Checked long-line clipping behavior

> Suggestion: keep the comments pane width fixed in split mode.

Inline code check: ` + "`make seed`" + ` and ` + "`make run`" + `.`
}

func seedMarkdownShowcaseCommentTwo() string {
	return `## QA Checklist

1. Create a task with multiline description.
2. Open viewer and scroll with j/k.
3. Verify due date coloring logic.

~~~bash
make reset-db
make seed
make run
~~~

| Case | Expected |
| --- | --- |
| h2 | emphasized |
| list | wrapped cleanly |
| code | monospaced block |
`
}
