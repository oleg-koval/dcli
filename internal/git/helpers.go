package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepo checks if the current directory is a Git repository
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// FetchOrigin fetches updates from the remote origin
func FetchOrigin(path string) error {
	cmd := exec.Command("git", "-C", path, "fetch", "origin") // #nosec G204 -- fixed command, no shell interpolation
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from origin: %w", err)
	}
	return nil
}

// CheckoutBranch checks out a branch in the repository
func CheckoutBranch(path string, branch string) error {
	cmd := exec.Command("git", "-C", path, "checkout", branch) // #nosec G204 -- branch is passed directly to git without shell expansion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}
	return nil
}

// ResetHard performs a hard reset to the specified commit/branch
func ResetHard(path string, target string) error {
	cmd := exec.Command("git", "-C", path, "reset", "--hard", target) // #nosec G204 -- target is passed directly to git without shell expansion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset hard to %s: %w", target, err)
	}
	return nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD") // #nosec G204 -- fixed command, no shell interpolation
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL returns the URL of the remote origin
func GetRemoteURL(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "config", "--get", "remote.origin.url") // #nosec G204 -- fixed command, no shell interpolation
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ValidateBranchTarget validates that a branch name is allowed for reset operations.
func ValidateBranchTarget(branch string) error {
	if strings.TrimSpace(branch) == "" {
		return fmt.Errorf("branch name cannot be empty or whitespace")
	}
	if branch != "develop" && branch != "acceptance" {
		return fmt.Errorf("branch must be 'develop' or 'acceptance', got '%s'", branch)
	}
	return nil
}
