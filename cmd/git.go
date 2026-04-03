package cmd

import (
	"github.com/oleg-koval/dcli/internal/git"
	"github.com/spf13/cobra"
)

// GitHelper defines the interface for Git operations
type GitHelper interface {
	IsGitRepo(path string) bool
	CheckoutBranch(path, branch string) error
	ResetHard(path, branch string) error
	FetchOrigin(path string) error
}

// Global helper - will be overridden in tests
var gitHelper GitHelper = &defaultGitHelper{}

type defaultGitHelper struct{}

func (g *defaultGitHelper) IsGitRepo(path string) bool {
	return git.IsGitRepo(path)
}

func (g *defaultGitHelper) CheckoutBranch(path, branch string) error {
	return git.CheckoutBranch(path, branch)
}

func (g *defaultGitHelper) ResetHard(path, branch string) error {
	return git.ResetHard(path, branch)
}

func (g *defaultGitHelper) FetchOrigin(path string) error {
	return git.FetchOrigin(path)
}

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git repository management commands",
}

func init() {
	rootCmd.AddCommand(gitCmd)
	gitCmd.AddCommand(gitResetCmd)
}
