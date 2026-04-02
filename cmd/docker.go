package cmd

import "github.com/spf13/cobra"

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker Compose management commands",
}

func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
}
