package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
)

func newCommentCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "comment",
		Short: "Comment operations",
	}
	c.AddCommand(newCommentListCommand())
	return c
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
