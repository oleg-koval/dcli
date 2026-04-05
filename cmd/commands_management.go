package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

var commandsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		return printCommandList(cmd, workspace)
	},
}

var commandsShowCmd = &cobra.Command{
	Use:   "show [path...]",
	Short: "Show one command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		item, err := resolveCommand(workspace, args)
		if err != nil {
			return err
		}
		return printCommandDetails(cmd.OutOrStdout(), item)
	},
}

var commandsAddCmd = &cobra.Command{
	Use:   "add [path...] -- <command...>",
	Short: "Add a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, step, err := splitPathAndStep(cmd, args)
		if err != nil {
			return err
		}

		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}

		command := buildSingleStepCommand(path, commandDescription, step, addAsShell)
		if err := workspace.AddCommand(command, scopeFromFlags(addAsShared)); err != nil {
			return err
		}
		if err := saveWorkspaceScope(workspace, addAsShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %s to your %s pack.\n", strings.Join(path, " "), packLabel(addAsShared))
		return nil
	},
}

var commandsEditCmd = &cobra.Command{
	Use:   "edit [path...]",
	Short: "Edit a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}

		path := append([]string(nil), args...)
		if dash := cmd.ArgsLenAtDash(); dash >= 0 {
			if dash == 0 {
				return errors.New("command path is required before --")
			}
			if dash >= len(args) {
				return errors.New("step command is required after --")
			}
			path = append([]string(nil), args[:dash]...)
		}

		var step []string
		if dash := cmd.ArgsLenAtDash(); dash >= 0 && dash < len(args) {
			step = append([]string(nil), args[dash:]...)
		}

		if err := workspace.UpdateCommand(path, scopeFromFlags(editShared), func(existing *commands.Command) error {
			if commandDescription != "" {
				existing.Description = commandDescription
			}
			if editEnable {
				existing.Enabled = true
			}
			if editDisable {
				existing.Enabled = false
			}
			if len(step) > 0 {
				existing.Steps = []commands.Step{{Type: commands.StepTypeExec, Command: append([]string(nil), step...)}}
			}
			existing.Revision++
			existing.UpdatedAt = time.Now().UTC()
			return nil
		}); err != nil {
			return err
		}
		if err := saveWorkspaceScope(workspace, editShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated %s in your %s pack.\n", strings.Join(path, " "), packLabel(editShared))
		return nil
	},
}

var commandsEnableCmd = &cobra.Command{
	Use:   "enable [path...]",
	Short: "Enable a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		if err := workspace.UpdateCommand(args, scopeFromFlags(enableShared), func(command *commands.Command) error {
			command.Enabled = true
			command.Revision++
			command.UpdatedAt = time.Now().UTC()
			return nil
		}); err != nil {
			return err
		}
		if err := saveWorkspaceScope(workspace, enableShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled %s in your %s pack.\n", strings.Join(args, " "), packLabel(enableShared))
		return nil
	},
}

var commandsDisableCmd = &cobra.Command{
	Use:   "disable [path...]",
	Short: "Disable a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		if err := workspace.UpdateCommand(args, scopeFromFlags(disableShared), func(command *commands.Command) error {
			command.Enabled = false
			command.Revision++
			command.UpdatedAt = time.Now().UTC()
			return nil
		}); err != nil {
			return err
		}
		if err := saveWorkspaceScope(workspace, disableShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled %s in your %s pack.\n", strings.Join(args, " "), packLabel(disableShared))
		return nil
	},
}

var commandsDeleteCmd = &cobra.Command{
	Use:   "delete [path...]",
	Short: "Delete a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}
		if !workspace.DeleteCommand(args, scopeFromFlags(deleteShared)) {
			return os.ErrNotExist
		}
		if err := saveWorkspaceScope(workspace, deleteShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from your %s pack.\n", strings.Join(args, " "), packLabel(deleteShared))
		return nil
	},
}

var commandsImportCmd = &cobra.Command{
	Use:   "import --file <path>",
	Short: "Import a command pack",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(importFile) == "" {
			return errors.New("--file is required")
		}

		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}

		pack, err := commands.LoadPackFile(importFile)
		if err != nil {
			return err
		}
		for _, item := range pack.Commands {
			if err := workspace.AddCommand(item, scopeFromFlags(importShared)); err != nil {
				return err
			}
		}
		if err := saveWorkspaceScope(workspace, importShared); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d command(s) from %s into your %s pack.\n", len(pack.Commands), importFile, packLabel(importShared))
		return nil
	},
}

var commandsExportCmd = &cobra.Command{
	Use:   "export --file <path>",
	Short: "Export a command pack",
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace, err := currentWorkspace()
		if err != nil {
			return err
		}

		pack := exportPack(workspace, exportShared)
		if strings.TrimSpace(exportFile) == "" {
			return writePack(cmd.OutOrStdout(), pack)
		}
		if err := pack.Save(exportFile); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d command(s) to %s.\n", len(pack.Commands), exportFile)
		return nil
	},
}

func init() {
	commandsAddCmd.Flags().StringVar(&commandDescription, "description", "", "Command description")
	commandsAddCmd.Flags().BoolVar(&addAsShared, "shared", false, "Store the command in the shared repo pack")
	commandsAddCmd.Flags().BoolVar(&addAsShell, "shell", false, "Store the step as a shell script")

	commandsEditCmd.Flags().StringVar(&commandDescription, "description", "", "Command description")
	commandsEditCmd.Flags().BoolVar(&editShared, "shared", false, "Edit the shared repo pack")
	commandsEditCmd.Flags().BoolVar(&editEnable, "enable", false, "Enable the command")
	commandsEditCmd.Flags().BoolVar(&editDisable, "disable", false, "Disable the command")
	commandsEditCmd.MarkFlagsMutuallyExclusive("enable", "disable")

	commandsEnableCmd.Flags().BoolVar(&enableShared, "shared", false, "Update the shared repo pack")
	commandsDisableCmd.Flags().BoolVar(&disableShared, "shared", false, "Update the shared repo pack")
	commandsDeleteCmd.Flags().BoolVar(&deleteShared, "shared", false, "Delete from the shared repo pack")

	commandsImportCmd.Flags().StringVar(&importFile, "file", "", "Import pack file")
	commandsImportCmd.Flags().BoolVar(&importShared, "shared", false, "Import into the shared repo pack")
	_ = commandsImportCmd.MarkFlagRequired("file")

	commandsExportCmd.Flags().StringVar(&exportFile, "file", "", "Export pack file")
	commandsExportCmd.Flags().BoolVar(&exportShared, "shared", false, "Export the shared repo pack only")
}

var (
	commandDescription string
	addAsShared        bool
	addAsShell         bool
	editShared         bool
	editEnable         bool
	editDisable        bool
	enableShared       bool
	disableShared      bool
	deleteShared       bool
	importFile         string
	importShared       bool
	exportFile         string
	exportShared       bool
)

func resolveCommand(workspace *commands.Workspace, path []string) (commands.ResolvedCommand, error) {
	for _, item := range workspace.ResolvedCommands(currentBuiltinPaths()) {
		if strings.EqualFold(item.Command.Key(), strings.Join(path, " ")) {
			return item, nil
		}
	}
	return commands.ResolvedCommand{}, os.ErrNotExist
}

func saveWorkspaceScope(workspace *commands.Workspace, shared bool) error {
	if shared {
		return workspace.SaveRepo()
	}
	return workspace.SaveLocal()
}

func writePack(w io.Writer, pack commands.Pack) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pack)
}

func exportPack(workspace *commands.Workspace, shared bool) commands.Pack {
	if shared {
		return workspace.Repo
	}
	pack := commands.Pack{Version: commands.PackVersion}
	pack.Commands = append(pack.Commands, workspace.Repo.Commands...)
	pack.Commands = append(pack.Commands, workspace.Local.Commands...)
	return pack
}

func printCommandDetails(w io.Writer, item commands.ResolvedCommand) error {
	if _, err := fmt.Fprintf(w, "PATH: %s\n", item.Command.DisplayName()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "STATUS: %s\n", item.Status); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "SCOPE: %s\n", item.Command.Scope); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "SOURCE: %s\n", item.Command.Source); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "DESCRIPTION: %s\n", item.Command.Description); err != nil {
		return err
	}
	for i, step := range item.Command.Steps {
		if _, err := fmt.Fprintf(w, "STEP %d: %s\n", i+1, step.Type); err != nil {
			return err
		}
	}
	return nil
}

func packLabel(shared bool) string {
	if shared {
		return "shared"
	}
	return "local"
}
