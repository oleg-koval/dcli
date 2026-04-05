package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List and manage custom command packs",
	Long: `List loaded commands by default, or use subcommands to manage custom automation packs.

Execution stays separate from management: custom commands run like normal dcli subcommands, while this command group handles browsing, editing, import, export, and sharing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		return printCommandList(cmd, workspace)
	},
}

func init() {
	rootCmd.AddCommand(commandsCmd)
	commandsCmd.AddCommand(commandsListCmd)
	commandsCmd.AddCommand(commandsShowCmd)
	commandsCmd.AddCommand(commandsAddCmd)
	commandsCmd.AddCommand(commandsEditCmd)
	commandsCmd.AddCommand(commandsEnableCmd)
	commandsCmd.AddCommand(commandsDisableCmd)
	commandsCmd.AddCommand(commandsDeleteCmd)
	commandsCmd.AddCommand(commandsImportCmd)
	commandsCmd.AddCommand(commandsExportCmd)
}

func currentWorkspace() (*commands.Workspace, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}
	return commands.LoadWorkspace(cwd)
}

func printCommandList(cmd *cobra.Command, workspace *commands.Workspace) error {
	return printResolvedCommands(cmd.OutOrStdout(), workspace.ResolvedCommands(currentBuiltinPaths()))
}

func printResolvedCommands(w io.Writer, resolved []commands.ResolvedCommand) error {
	sorted := make([]commands.ResolvedCommand, len(resolved))
	copy(sorted, resolved)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Command.Key() < sorted[j].Command.Key()
	})

	if _, err := fmt.Fprintln(w, "PATH\tSTATUS\tSCOPE\tSOURCE\tDESCRIPTION"); err != nil {
		return err
	}
	for _, item := range sorted {
		if _, err := fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\n",
			item.Command.DisplayName(),
			item.Status,
			item.Command.Scope,
			item.Command.Source,
			strings.TrimSpace(item.Command.Description),
		); err != nil {
			return err
		}
	}
	return nil
}
