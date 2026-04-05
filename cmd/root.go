package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/autoupdate"
	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:           "dcli",
	Short:         "Developer CLI for Docker, Git, and command execution",
	Long:          `dcli is an execution-first command-line tool for Docker services, Git workflows, and user-defined command automation. Use dcli commands to manage custom packs and dcli commands ui for the interactive browser.`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
		}
	},
}

type startupUpdater interface {
	Run(ctx context.Context, currentVersion string, args []string)
}

var autoUpdateRunner startupUpdater = autoupdate.NewRunner(autoupdate.Repository{
	Owner: "oleg-koval",
	Name:  "dcli",
})

func init() {
	// Note: dockerCmd is added in docker.go init function
}

// Execute runs the root command
func Execute() {
	autoUpdateRunner.Run(context.Background(), Version, os.Args)

	if err := registerCustomCommands(rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// GetRootCmd returns the root command (useful for testing)
func GetRootCmd() *cobra.Command {
	return rootCmd
}
