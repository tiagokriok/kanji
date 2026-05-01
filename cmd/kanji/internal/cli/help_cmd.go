package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newHelpCommand() *cobra.Command {
	help := &cobra.Command{
		Use:   "help [topic]",
		Short: "Help about kanji topics and commands",
		Long: `Help provides detailed guidance on kanji concepts, context model,
selectors, and output formats.

Available topics:
  concepts    Core kanji concepts and hierarchy
  context     Namespace and context model
  selectors   Resource selection rules
  output      Output formats and options
`,
	}

	help.AddCommand(newHelpTopicCommand("concepts", helpConceptsLong))
	help.AddCommand(newHelpTopicCommand("context", helpContextLong))
	help.AddCommand(newHelpTopicCommand("selectors", helpSelectorsLong))
	help.AddCommand(newHelpTopicCommand("output", helpOutputLong))

	return help
}

func newHelpTopicCommand(name, long string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Help about %s", name),
		Long:  long,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), long)
			return nil
		},
	}
}

const helpConceptsLong = `Concepts

kanji organizes work in a hierarchy:

  Provider -> Workspace -> Board -> Column -> Task -> Comment

Provider
  The identity layer. Each kanji installation has at least one local provider.

Workspace
  A container for boards. Workspace names are unique globally.

Board
  A container for columns within a workspace. Board names are unique within a workspace.

Column
  A stage in a workflow. Columns belong to a board and have a WIP limit. Column names are unique within a board.

Task
  A work item with title, description, priority, due date, labels, and status. Status is driven by the column the task is in.

Comment
  A note attached to a task with an author and body.

Bootstrap
  Before using kanji, run 'kanji data bootstrap' to create the default provider, workspace, board, and columns.
`

const helpContextLong = `Context

kanji uses a namespace model based on the current working directory.

Namespace Resolution
  By default, the namespace key is the exact normalized cwd.
  Set KANJI_CONTEXT to override it globally.

CLI Context (cli_context)
  The explicit workspace and board selected via 'kanji context set'.
  Some commands can infer scope from cli_context when flags are omitted.

TUI State (tui_state)
  The workspace and board last used in the TUI. Stored alongside cli_context
  in the shared namespaced state file, but the CLI does not modify it.

Context Commands
  kanji context show    Display current namespace and context
  kanji context set     Set explicit workspace and/or board
  kanji context clear   Clear explicit CLI context only
`

const helpSelectorsLong = `Selectors

Resources can be selected by ID or by exact normalized name.

ID Selectors
  --workspace-id, --board-id, --column-id, --task-id, --comment-id

Name Selectors
  --workspace, --board, --column, --task

Matching Rules
  - Names are matched case-insensitively.
  - Surrounding whitespace is trimmed.
  - Exact match only; partial matches are rejected as ambiguous.

Scope Requirements
  Board names require a workspace scope.
  Column names require a board scope.
  Task titles require a workspace scope and optionally a board scope.
`

const helpOutputLong = `Output

kanji supports two output modes.

Human Output (default)
  List commands print tables.
  Get commands print key/value blocks.
  Descriptions and comments render clearly in detail views.

JSON Output (--json)
  All responses are wrapped.
  Lists are wrapped with the resource name and include counts.
  Singular objects are wrapped with the resource name.
  Errors use a standard { error: { code, message } } shape.

Verbose Output (--verbose)
  Adds runtime metadata such as namespace key, db path, and resolution details.
  Does not alter the JSON contract.
`
