package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/application"
)

func newTaskCommand() *cobra.Command {
	t := &cobra.Command{
		Use:   "task",
		Short: "Task operations",
	}
	t.AddCommand(newTaskListCommand())
	return t
}

func newTaskListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks for a workspace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runTaskList(cmd, ns)
		},
	}
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().String("board-id", "", "board ID (optional narrowing)")
	cmd.Flags().String("board", "", "board name (optional narrowing)")
	cmd.Flags().String("query", "", "title query filter")
	cmd.Flags().String("column", "", "column ID filter")
	cmd.Flags().Int("due-soon", 0, "due within N days")
	return cmd
}

func runTaskList(cmd *cobra.Command, ns Namespace) error {
	cfg, err := ResolveConfig(cmd)
	if err != nil {
		return err
	}

	rt, err := NewRuntime(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer rt.Close()

	if err := GuardBootstrap(rt); err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve workspace scope.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") {
		workspaceID, _ = cmd.Flags().GetString("workspace-id")
	} else if cmd.Flags().Changed("workspace") {
		name, _ := cmd.Flags().GetString("workspace")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return err
		}
		found := false
		for _, ws := range workspaces {
			if ExactMatch(ws.Name, name) {
				workspaceID = ws.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("workspace", name)
		}
	} else {
		return NewValidation("workspace scope required: use --workspace-id, --workspace, or kanji context set")
	}

	// Resolve optional board narrowing.
	var boardID string
	if cmd.Flags().Changed("board-id") {
		boardID, _ = cmd.Flags().GetString("board-id")
	} else if cmd.Flags().Changed("board") {
		name, _ := cmd.Flags().GetString("board")
		boards, err := rt.ContextService.ListBoards(ctx, workspaceID)
		if err != nil {
			return err
		}
		found := false
		for _, b := range boards {
			if ExactMatch(b.Name, name) {
				boardID = b.ID
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("board", name)
		}
	}

	filters := application.ListTaskFilters{
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}

	if cmd.Flags().Changed("query") {
		filters.TitleQuery, _ = cmd.Flags().GetString("query")
	}
	if cmd.Flags().Changed("column") {
		filters.ColumnID, _ = cmd.Flags().GetString("column")
	}
	if cmd.Flags().Changed("due-soon") {
		filters.DueSoonDays, _ = cmd.Flags().GetInt("due-soon")
	}

	tasks, err := rt.TaskFlow.ListTasks(ctx, filters)
	if err != nil {
		return err
	}

	if cfg.JSON {
		items := make([]map[string]string, len(tasks))
		for i, task := range tasks {
			status := ""
			if task.Status != nil {
				status = *task.Status
			}
			items[i] = map[string]string{
				"id":       task.ID,
				"title":    task.Title,
				"status":   status,
				"priority": strconv.Itoa(task.Priority),
			}
		}
		return RenderWrappedListJSON(cmd.OutOrStdout(), "tasks", items, len(tasks))
	}

	headers := []string{"ID", "Title", "Status", "Priority"}
	rows := make([][]string, len(tasks))
	for i, task := range tasks {
		status := ""
		if task.Status != nil {
			status = *task.Status
		}
		rows[i] = []string{task.ID, task.Title, status, strconv.Itoa(task.Priority)}
	}
	return RenderTable(cmd.OutOrStdout(), headers, rows)
}
