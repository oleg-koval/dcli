package cmd

import "github.com/spf13/cobra"

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git repository management commands",
}

func init() {
	rootCmd.AddCommand(gitCmd)
	gitCmd.AddCommand(gitResetCmd)
}
