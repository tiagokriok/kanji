package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/application"
)

// AssembleCreateTaskInput builds a CreateTaskInput from command flags.
// It resolves title, description, priority, due date, and labels.
// ProviderID and Status must be set by the caller.
func AssembleCreateTaskInput(cmd *cobra.Command, workspaceID, boardID, columnID string) (application.CreateTaskInput, error) {
	title, err := cmd.Flags().GetString("title")
	if err != nil {
		return application.CreateTaskInput{}, NewValidation("read --title: " + err.Error())
	}
	if strings.TrimSpace(title) == "" {
		return application.CreateTaskInput{}, NewValidation("title is required")
	}

	desc, err := ResolveTextInput(cmd, "description", "description-file", true, nil)
	if err != nil {
		return application.CreateTaskInput{}, err
	}

	priority := 3 // default medium
	if cmd.Flags().Changed("priority") {
		pStr, _ := cmd.Flags().GetString("priority")
		p, err := ParsePriority(pStr)
		if err != nil {
			return application.CreateTaskInput{}, err
		}
		priority = p
	}

	var dueAt *time.Time
	if cmd.Flags().Changed("due-date") {
		dStr, _ := cmd.Flags().GetString("due-date")
		d, err := ParseDueDate(dStr)
		if err != nil {
			return application.CreateTaskInput{}, err
		}
		dueAt = &d
	}

	var labels []string
	if cmd.Flags().Changed("labels") {
		l, _ := cmd.Flags().GetStringSlice("labels")
		labels = NormalizeLabels(l)
	}

	return application.CreateTaskInput{
		WorkspaceID:   workspaceID,
		BoardID:       &boardID,
		ColumnID:      &columnID,
		Title:         title,
		DescriptionMD: desc,
		Priority:      priority,
		DueAt:         dueAt,
		Labels:        labels,
	}, nil
}

// AssembleUpdateTaskInput builds an UpdateTaskInput with only the fields
// that were changed on the command. It supports clear flags for
// description, due date, and labels.
func AssembleUpdateTaskInput(cmd *cobra.Command) (application.UpdateTaskInput, error) {
	var input application.UpdateTaskInput

	if cmd.Flags().Changed("title") {
		t, _ := cmd.Flags().GetString("title")
		input.Title = &t
	}

	descChanged := cmd.Flags().Changed("description") || cmd.Flags().Changed("description-file")
	clearDesc := cmd.Flags().Changed("clear-description")
	if descChanged && clearDesc {
		return application.UpdateTaskInput{}, NewValidation("--description / --description-file and --clear-description are mutually exclusive")
	}
	if descChanged {
		d, err := ResolveTextInput(cmd, "description", "description-file", true, nil)
		if err != nil {
			return application.UpdateTaskInput{}, err
		}
		input.DescriptionMD = &d
	}
	if clearDesc {
		empty := ""
		input.DescriptionMD = &empty
	}

	if cmd.Flags().Changed("priority") {
		pStr, _ := cmd.Flags().GetString("priority")
		p, err := ParsePriority(pStr)
		if err != nil {
			return application.UpdateTaskInput{}, err
		}
		input.Priority = &p
	}

	dueChanged := cmd.Flags().Changed("due-date")
	clearDue := cmd.Flags().Changed("clear-due-date")
	if dueChanged && clearDue {
		return application.UpdateTaskInput{}, NewValidation("--due-date and --clear-due-date are mutually exclusive")
	}
	if dueChanged {
		dStr, _ := cmd.Flags().GetString("due-date")
		d, err := ParseDueDate(dStr)
		if err != nil {
			return application.UpdateTaskInput{}, err
		}
		input.DueAt = &d
	}
	if clearDue {
		zero := time.Time{}
		input.DueAt = &zero
	}

	labelsChanged := cmd.Flags().Changed("labels")
	clearLabels := cmd.Flags().Changed("clear-labels")
	if labelsChanged && clearLabels {
		return application.UpdateTaskInput{}, NewValidation("--labels and --clear-labels are mutually exclusive")
	}
	if labelsChanged {
		l, _ := cmd.Flags().GetStringSlice("labels")
		normalized := NormalizeLabels(l)
		input.Labels = &normalized
	}
	if clearLabels {
		empty := []string{}
		input.Labels = &empty
	}

	return input, nil
}

// ResolveTaskID resolves a task ID from --task-id or --task flags.
// When --task is used, workspaceID must be provided for title resolution.
func ResolveTaskID(cmd *cobra.Command, rt *Runtime, workspaceID string) (string, error) {
	ctx := context.Background()

	if cmd.Flags().Changed("task-id") {
		id, _ := cmd.Flags().GetString("task-id")
		workspaces, err := rt.ContextService.ListWorkspaces(ctx)
		if err != nil {
			return "", err
		}
		for _, ws := range workspaces {
			filters := application.ListTaskFilters{WorkspaceID: ws.ID}
			tasks, err := rt.TaskFlow.ListTasks(ctx, filters)
			if err != nil {
				return "", err
			}
			for _, t := range tasks {
				if t.ID == id {
					return id, nil
				}
			}
		}
		return "", NewNotFound("task", id)
	}

	if cmd.Flags().Changed("task") {
		title, _ := cmd.Flags().GetString("task")
		if workspaceID == "" {
			return "", NewValidation("workspace scope required for task title resolution")
		}
		filters := application.ListTaskFilters{WorkspaceID: workspaceID}
		tasks, err := rt.TaskFlow.ListTasks(ctx, filters)
		if err != nil {
			return "", err
		}
		for _, t := range tasks {
			if ExactMatch(t.Title, title) {
				return t.ID, nil
			}
		}
		return "", NewNotFound("task", title)
	}

	return "", NewValidation("task-id or task is required")
}

// ResolveMoveDestination resolves a destination column ID and status from
// --to-column-id or --to-column flags within the given board.
func ResolveMoveDestination(cmd *cobra.Command, rt *Runtime, boardID string) (string, string, error) {
	ctx := context.Background()

	columns, err := rt.ContextService.ListColumns(ctx, boardID)
	if err != nil {
		return "", "", err
	}

	if cmd.Flags().Changed("to-column-id") {
		id, _ := cmd.Flags().GetString("to-column-id")
		for _, col := range columns {
			if col.ID == id {
				return col.ID, strings.ToLower(col.Name), nil
			}
		}
		return "", "", NewNotFound("column", id)
	}

	if cmd.Flags().Changed("to-column") {
		name, _ := cmd.Flags().GetString("to-column")
		for _, col := range columns {
			if ExactMatch(col.Name, name) {
				return col.ID, strings.ToLower(col.Name), nil
			}
		}
		return "", "", NewNotFound("column", name)
	}

	return "", "", NewValidation("to-column-id or to-column is required")
}

// ── Commands ──

func newTaskCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runTaskCreate(cmd, ns)
		},
	}
	cmd.Flags().String("title", "", "task title (required)")
	cmd.Flags().String("description", "", "task description")
	cmd.Flags().String("description-file", "", "read description from file (use - for stdin)")
	cmd.Flags().String("priority", "", "priority: critical, urgent, high, medium, low, none, or 0-5")
	cmd.Flags().String("due-date", "", "due date: YYYY-MM-DD or RFC3339")
	cmd.Flags().StringSlice("labels", nil, "comma-separated labels")
	cmd.Flags().String("workspace-id", "", "workspace ID")
	cmd.Flags().String("workspace", "", "workspace name")
	cmd.Flags().String("board-id", "", "board ID")
	cmd.Flags().String("board", "", "board name")
	cmd.Flags().String("column-id", "", "column ID")
	cmd.Flags().String("column", "", "column name")
	return cmd
}

func runTaskCreate(cmd *cobra.Command, ns Namespace) error {
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

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve workspace.
	workspaceID, _, err := ResolveWorkspaceScope(cmd, rt, store, ns)
	if err != nil {
		return err
	}

	// Get provider ID from workspace.
	workspaces, err := rt.ContextService.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	var providerID string
	for _, ws := range workspaces {
		if ws.ID == workspaceID {
			providerID = ws.ProviderID
			break
		}
	}
	if providerID == "" {
		return NewNotFound("workspace", workspaceID)
	}

	// Resolve board.
	boardID, _, err := ResolveBoardScope(cmd, rt, store, ns, workspaceID)
	if err != nil {
		return err
	}

	// Resolve column.
	var columnID, status string
	if cmd.Flags().Changed("column-id") {
		columnID, _ = cmd.Flags().GetString("column-id")
		columns, err := rt.ContextService.ListColumns(ctx, boardID)
		if err != nil {
			return err
		}
		found := false
		for _, col := range columns {
			if col.ID == columnID {
				status = strings.ToLower(col.Name)
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("column", columnID)
		}
	} else if cmd.Flags().Changed("column") {
		name, _ := cmd.Flags().GetString("column")
		columns, err := rt.ContextService.ListColumns(ctx, boardID)
		if err != nil {
			return err
		}
		found := false
		for _, col := range columns {
			if ExactMatch(col.Name, name) {
				columnID = col.ID
				status = strings.ToLower(col.Name)
				found = true
				break
			}
		}
		if !found {
			return NewNotFound("column", name)
		}
	} else {
		// Default to first column.
		columns, err := rt.ContextService.ListColumns(ctx, boardID)
		if err != nil {
			return err
		}
		if len(columns) == 0 {
			return NewValidation("board has no columns")
		}
		columnID = columns[0].ID
		status = strings.ToLower(columns[0].Name)
	}

	input, err := AssembleCreateTaskInput(cmd, workspaceID, boardID, columnID)
	if err != nil {
		return err
	}
	input.ProviderID = providerID
	input.Status = &status

	task, err := rt.TaskService.CreateTask(ctx, input)
	if err != nil {
		return err
	}

	if cfg.JSON {
		data := map[string]interface{}{
			"id":       task.ID,
			"title":    task.Title,
			"priority": task.Priority,
		}
		if task.Status != nil {
			data["status"] = *task.Status
		}
		return RenderWriteResultJSON(cmd.OutOrStdout(), "task", data)
	}

	fields := map[string]string{
		"Title":    task.Title,
		"Priority": strconv.Itoa(task.Priority),
	}
	if task.Status != nil {
		fields["Status"] = *task.Status
	}
	return RenderWriteResult(cmd.OutOrStdout(), "Task", task.ID, fields)
}

func newTaskUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runTaskUpdate(cmd, ns)
		},
	}
	cmd.Flags().String("task-id", "", "task ID")
	cmd.Flags().String("task", "", "task title")
	cmd.Flags().String("title", "", "new title")
	cmd.Flags().String("description", "", "new description")
	cmd.Flags().String("description-file", "", "read description from file (use - for stdin)")
	cmd.Flags().String("priority", "", "new priority")
	cmd.Flags().String("due-date", "", "new due date")
	cmd.Flags().StringSlice("labels", nil, "new labels")
	cmd.Flags().Bool("clear-description", false, "clear description")
	cmd.Flags().Bool("clear-due-date", false, "clear due date")
	cmd.Flags().Bool("clear-labels", false, "clear labels")
	cmd.Flags().String("workspace-id", "", "workspace ID (required for title resolution)")
	cmd.Flags().String("workspace", "", "workspace name (required for title resolution)")
	return cmd
}

func runTaskUpdate(cmd *cobra.Command, ns Namespace) error {
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

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve workspace for task title resolution.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") || cmd.Flags().Changed("workspace") {
		workspaceID, _, err = ResolveWorkspaceScope(cmd, rt, store, ns)
		if err != nil {
			return err
		}
	}

	taskID, err := ResolveTaskID(cmd, rt, workspaceID)
	if err != nil {
		return err
	}

	input, err := AssembleUpdateTaskInput(cmd)
	if err != nil {
		return err
	}

	// Ensure at least one patch field is present.
	if input.Title == nil && input.DescriptionMD == nil && input.Priority == nil &&
		input.DueAt == nil && input.Labels == nil {
		return NewValidation("at least one of --title, --description, --priority, --due-date, --labels, --clear-description, --clear-due-date, --clear-labels is required")
	}

	if err := rt.TaskService.UpdateTask(ctx, taskID, input); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "task", map[string]interface{}{
			"id":      taskID,
			"updated": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Task updated\nID:  %s\n", taskID)
	return nil
}

func newTaskMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a task to a different column",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runTaskMove(cmd, ns)
		},
	}
	cmd.Flags().String("task-id", "", "task ID")
	cmd.Flags().String("task", "", "task title")
	cmd.Flags().String("to-column-id", "", "destination column ID")
	cmd.Flags().String("to-column", "", "destination column name")
	cmd.Flags().String("workspace-id", "", "workspace ID (required for title resolution)")
	cmd.Flags().String("workspace", "", "workspace name (required for title resolution)")
	cmd.Flags().String("board-id", "", "board ID (required for column name resolution)")
	cmd.Flags().String("board", "", "board name (required for column name resolution)")
	return cmd
}

func runTaskMove(cmd *cobra.Command, ns Namespace) error {
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

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve workspace for task title resolution.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") || cmd.Flags().Changed("workspace") {
		workspaceID, _, err = ResolveWorkspaceScope(cmd, rt, store, ns)
		if err != nil {
			return err
		}
	}

	taskID, err := ResolveTaskID(cmd, rt, workspaceID)
	if err != nil {
		return err
	}

	// Resolve board for column resolution.
	// If task was found by ID, we need its board for column validation.
	// Look up the task to get its board ID.
	task, err := rt.TaskService.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	boardID := ""
	if task.BoardID != nil {
		boardID = *task.BoardID
	}
	if boardID == "" {
		return NewValidation("task has no board")
	}

	// If --to-column or --to-column-id uses name resolution, board scope is needed.
	if cmd.Flags().Changed("to-column") {
		boardID, _, err = ResolveBoardScope(cmd, rt, store, ns, task.WorkspaceID)
		if err != nil {
			return err
		}
	}

	columnID, status, err := ResolveMoveDestination(cmd, rt, boardID)
	if err != nil {
		return err
	}

	if err := rt.TaskFlow.MoveTask(ctx, taskID, &columnID, &status, 0); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "task", map[string]interface{}{
			"id":        taskID,
			"column_id": columnID,
			"status":    status,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Task moved\nID:      %s\nColumn:  %s\nStatus:  %s\n", taskID, columnID, status)
	return nil
}

func newTaskDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runTaskDelete(cmd, ns)
		},
	}
	cmd.Flags().String("task-id", "", "task ID")
	cmd.Flags().String("task", "", "task title")
	cmd.Flags().Bool("yes", false, "confirm deletion")
	cmd.Flags().String("workspace-id", "", "workspace ID (required for title resolution)")
	cmd.Flags().String("workspace", "", "workspace name (required for title resolution)")
	return cmd
}

func runTaskDelete(cmd *cobra.Command, ns Namespace) error {
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

	if err := RequireConfirmation(cmd, "yes"); err != nil {
		return err
	}

	store, err := defaultStateStore()
	if err != nil {
		return err
	}

	// Resolve workspace for task title resolution.
	var workspaceID string
	if cmd.Flags().Changed("workspace-id") || cmd.Flags().Changed("workspace") {
		workspaceID, _, err = ResolveWorkspaceScope(cmd, rt, store, ns)
		if err != nil {
			return err
		}
	}

	taskID, err := ResolveTaskID(cmd, rt, workspaceID)
	if err != nil {
		return err
	}

	if err := rt.TaskService.DeleteTask(context.Background(), taskID); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "task", map[string]interface{}{
			"id":      taskID,
			"deleted": true,
		})
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Task deleted\nID:  %s\n", taskID)
	return nil
}
