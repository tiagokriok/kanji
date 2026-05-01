package cli

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tiagokriok/kanji/internal/application"
	"github.com/tiagokriok/kanji/internal/domain"
)

func newCommentCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "comment",
		Short: "Comment operations",
	}
	c.AddCommand(newCommentListCommand())
	c.AddCommand(newCommentGetCommand())
	c.AddCommand(newCommentCreateCommand())
	c.AddCommand(newCommentUpdateCommand())
	c.AddCommand(newCommentDeleteCommand())
	return c
}

func newCommentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runCommentCreate(cmd, ns)
		},
	}
	cmd.Flags().String("task-id", "", "task ID")
	cmd.Flags().String("body", "", "comment body")
	cmd.Flags().String("body-file", "", "path to file containing comment body (- for stdin)")
	cmd.Flags().String("author", "", "comment author")
	return cmd
}

func newCommentUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a comment body",
		Long: `Update a comment body. The body completely replaces the existing content.
Supports inline text, file input, or stdin (-).`,
		Example: `  kanji comment update --comment-id <id> --body "Updated comment"
  kanji comment update --comment-id <id> --body-file comment.md`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runCommentUpdate(cmd, ns)
		},
	}
	cmd.Flags().String("comment-id", "", "comment ID")
	cmd.Flags().String("body", "", "new comment body")
	cmd.Flags().String("body-file", "", "path to file containing comment body (- for stdin)")
	return cmd
}

func newCommentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a comment",
		Long:    `Delete a comment permanently. Requires --yes for confirmation.`,
		Example: `  kanji comment delete --comment-id <id> --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runCommentDelete(cmd, ns)
		},
	}
	cmd.Flags().String("comment-id", "", "comment ID")
	cmd.Flags().Bool("yes", false, "confirm deletion")
	return cmd
}

func runCommentCreate(cmd *cobra.Command, ns Namespace) error {
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

	// Resolve task.
	var taskID string
	if cmd.Flags().Changed("task-id") {
		taskID, _ = cmd.Flags().GetString("task-id")
	} else {
		return NewValidation("task-id is required")
	}

	task, err := rt.TaskService.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewNotFound("task", taskID)
		}
		return err
	}

	// Resolve body.
	body, err := ResolveTextInput(cmd, "body", "body-file", false, nil)
	if err != nil {
		return err
	}

	// Optional author.
	var author *string
	if cmd.Flags().Changed("author") {
		a, _ := cmd.Flags().GetString("author")
		author = &a
	}

	comment, err := rt.CommentService.AddComment(ctx, application.AddCommentInput{
		TaskID:     task.ID,
		ProviderID: task.ProviderID,
		BodyMD:     body,
		Author:     author,
	})
	if err != nil {
		return err
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"id":      comment.ID,
			"task_id": comment.TaskID,
			"body":    comment.BodyMD,
		}
		if comment.Author != nil {
			payload["author"] = *comment.Author
		}
		return RenderWriteResultJSON(cmd.OutOrStdout(), "comment", payload)
	}

	fields := map[string]string{
		"Task ID": comment.TaskID,
		"Body":    comment.BodyMD,
	}
	if comment.Author != nil {
		fields["Author"] = *comment.Author
	}
	return RenderWriteResult(cmd.OutOrStdout(), "comment", comment.ID, fields)
}

func newCommentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments for a task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runCommentList(cmd, ns)
		},
	}
	cmd.Flags().String("task-id", "", "task ID")
	return cmd
}

func newCommentGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a comment by ID",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ns, err := ResolveNamespace()
			if err != nil {
				return err
			}
			return runCommentGet(cmd, ns)
		},
	}
	cmd.Flags().String("comment-id", "", "comment ID")
	return cmd
}

func runCommentUpdate(cmd *cobra.Command, ns Namespace) error {
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

	var commentID string
	if cmd.Flags().Changed("comment-id") {
		commentID, _ = cmd.Flags().GetString("comment-id")
	} else {
		return NewValidation("comment-id is required")
	}

	body, err := ResolveTextInput(cmd, "body", "body-file", false, nil)
	if err != nil {
		return err
	}

	if err := rt.CommentService.UpdateComment(ctx, commentID, body); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderWriteResultJSON(cmd.OutOrStdout(), "comment", map[string]interface{}{
			"id":   commentID,
			"body": body,
		})
	}

	_, _ = cmd.OutOrStdout().Write([]byte("comment updated\n"))
	return RenderKV(cmd.OutOrStdout(), map[string]string{
		"ID":   commentID,
		"Body": body,
	})
}

func runCommentDelete(cmd *cobra.Command, ns Namespace) error {
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

	var commentID string
	if cmd.Flags().Changed("comment-id") {
		commentID, _ = cmd.Flags().GetString("comment-id")
	} else {
		return NewValidation("comment-id is required")
	}

	if err := RequireConfirmation(cmd, "yes"); err != nil {
		return err
	}

	if err := rt.CommentService.DeleteComment(ctx, commentID); err != nil {
		return err
	}

	if cfg.JSON {
		return RenderDeleteResultJSON(cmd.OutOrStdout(), "comment", commentID, false)
	}

	return RenderDeleteResult(cmd.OutOrStdout(), "comment", commentID)
}

func runCommentGet(cmd *cobra.Command, ns Namespace) error {
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

	// Resolve comment.
	var commentID string
	if cmd.Flags().Changed("comment-id") {
		commentID, _ = cmd.Flags().GetString("comment-id")
	} else {
		return NewValidation("comment-id is required")
	}

	// Search across all tasks.
	workspaces, err := rt.ContextService.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	var comment domain.Comment
	found := false
	for _, ws := range workspaces {
		filters := application.ListTaskFilters{WorkspaceID: ws.ID}
		tasks, err := rt.TaskFlow.ListTasks(ctx, filters)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			comments, err := rt.CommentService.ListComments(ctx, task.ID)
			if err != nil {
				return err
			}
			for _, c := range comments {
				if c.ID == commentID {
					comment = c
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return NewNotFound("comment", commentID)
	}

	if cfg.JSON {
		payload := map[string]interface{}{
			"id":      comment.ID,
			"task_id": comment.TaskID,
			"body":    comment.BodyMD,
		}
		if comment.Author != nil {
			payload["author"] = *comment.Author
		}
		return RenderWrappedJSON(cmd.OutOrStdout(), "comment", payload)
	}

	pairs := map[string]string{
		"ID":      comment.ID,
		"Task ID": comment.TaskID,
		"Body":    comment.BodyMD,
	}
	if comment.Author != nil {
		pairs["Author"] = *comment.Author
	}
	return RenderKV(cmd.OutOrStdout(), pairs)
}

func runCommentList(cmd *cobra.Command, ns Namespace) error {
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

	// Resolve task scope.
	var taskID string
	if cmd.Flags().Changed("task-id") {
		taskID, _ = cmd.Flags().GetString("task-id")
	} else {
		return NewValidation("task-id is required")
	}

	comments, err := rt.CommentService.ListComments(ctx, taskID)
	if err != nil {
		return err
	}

	if cfg.JSON {
		items := make([]map[string]string, len(comments))
		for i, c := range comments {
			preview := c.BodyMD
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			author := ""
			if c.Author != nil {
				author = *c.Author
			}
			items[i] = map[string]string{
				"id":         c.ID,
				"task_id":    c.TaskID,
				"author":     author,
				"created_at": c.CreatedAt.Format("2006-01-02"),
				"preview":    preview,
			}
		}
		return RenderWrappedListJSON(cmd.OutOrStdout(), "comments", items, len(comments))
	}

	headers := []string{"ID", "Task ID", "Author", "Created", "Preview"}
	rows := make([][]string, len(comments))
	for i, c := range comments {
		preview := c.BodyMD
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		author := ""
		if c.Author != nil {
			author = *c.Author
		}
		rows[i] = []string{c.ID, c.TaskID, author, c.CreatedAt.Format("2006-01-02"), preview}
	}
	return RenderTable(cmd.OutOrStdout(), headers, rows)
}
