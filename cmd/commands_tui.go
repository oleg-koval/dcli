package cmd

import (
	"fmt"
	"io"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	commandsui "github.com/oleg-koval/dcli/internal/commands/ui"
	"github.com/spf13/cobra"
)

var commandBrowserRunner = func(model tea.Model, in io.Reader, out io.Writer) error {
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(in), tea.WithOutput(out))
	_, err := program.Run()
	return err
}

var commandsUICmd = &cobra.Command{
	Use:   "ui",
	Short: "Open the command management TUI",
	Long:  "Use the terminal UI to browse, organize, share, import, export, enable, disable, or delete command packs. This is a management surface, not the default execution path.",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}

		exportPath := commandsUIExportFile
		if exportPath == "" {
			exportPath = filepath.Join(workspace.Cwd, "dcli-commands-export.json")
		}

		model := commandsui.NewModel(workspace, currentBuiltinPaths(), exportPath)
		if err := commandBrowserRunner(model, cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("run command browser: %w", err)
		}
		return nil
	},
}

var commandsUIExportFile string

func init() {
	commandsUICmd.Flags().StringVar(&commandsUIExportFile, "export-file", "", "Path used for import/export actions in UI mode")
	commandsCmd.AddCommand(commandsUICmd)
}
