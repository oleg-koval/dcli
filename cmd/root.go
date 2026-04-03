package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:     "dcli",
	Short:   "Docker CLI - Manage Docker containers and services",
	Long:    `dcli is a command-line tool for managing Docker containers and services with Git integration support.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

func init() {
	// Note: dockerCmd is added in docker.go init function
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// GetRootCmd returns the root command (useful for testing)
func GetRootCmd() *cobra.Command {
	return rootCmd
}
