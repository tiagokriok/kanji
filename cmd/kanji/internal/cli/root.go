package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "kanji",
		Short: "kanji - task management CLI and TUI",
		Long: `kanji is a CLI-first task management tool with an optional TUI.

Before using resource commands, bootstrap the system:
  kanji data bootstrap

To launch the TUI:
  kanji tui

Use "kanji help" for available commands.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().String("db-path", "", "path to SQLite database (env: KANJI_DB_PATH)")
	root.PersistentFlags().Bool("json", false, "output in JSON format")
	root.PersistentFlags().Bool("verbose", false, "enable verbose output")

	root.AddCommand(newDataCommand())
	root.AddCommand(newHelpCommand())
	root.AddCommand(newContextCommand())
	root.AddCommand(newDBCommand())
	root.AddCommand(newWorkspaceCommand())
	root.AddCommand(newBoardCommand())

	return root
}

func Execute() error {
	return NewRootCommand().Execute()
}
