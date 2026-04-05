package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

var loadWorkspace = commands.LoadWorkspace
var builtinPathsSnapshot [][]string

func registerCustomCommands(root *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	workspace, err := loadWorkspace(cwd)
	if err != nil {
		return err
	}

	reserved := builtinCommandPaths(root)
	builtinPathsSnapshot = reserved
	active := workspace.ActiveCommands(reserved)
	for _, resolved := range active {
		if err := registerCommandPath(root, resolved.Command); err != nil {
			return err
		}
	}
	return nil
}

func registerCommandPath(root *cobra.Command, command commands.Command) error {
	current := root
	for index, segment := range command.Path {
		child := findOrCreateChild(current, segment)
		if index == len(command.Path)-1 {
			leaf := command.Clone()
			child.Short = leaf.Description
			child.Long = leaf.Description
			child.Use = segment
			child.Args = cobra.ArbitraryArgs
			child.RunE = func(cmd *cobra.Command, args []string) error {
				return commands.Execute(context.Background(), leaf, args, cmd.OutOrStdout(), cmd.ErrOrStderr())
			}
		}
		current = child
	}
	return nil
}

func findOrCreateChild(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}

	child := &cobra.Command{Use: name}
	parent.AddCommand(child)
	return child
}

func builtinCommandPaths(root *cobra.Command) [][]string {
	paths := make([][]string, 0)
	collectCommandPaths(root, nil, &paths)
	return paths
}

func currentBuiltinPaths() [][]string {
	if len(builtinPathsSnapshot) > 0 {
		return builtinPathsSnapshot
	}
	return builtinCommandPaths(rootCmd)
}

func collectCommandPaths(command *cobra.Command, prefix []string, paths *[][]string) {
	if command == nil {
		return
	}

	if command != rootCmd {
		current := append(append([]string(nil), prefix...), command.Name())
		if len(current) > 0 {
			*paths = append(*paths, current)
		}
		prefix = current
	}

	for _, child := range command.Commands() {
		collectCommandPaths(child, prefix, paths)
	}
}
