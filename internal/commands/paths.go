package commands

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	localPackFileName = "commands.json"
	repoPackDirName   = ".dcli"
)

// LocalPackPath returns the user-local command pack path.
func LocalPackPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, repoPackDirName, localPackFileName), nil
}

// RepoRoot walks upward until it finds a git worktree root.
func RepoRoot(start string) (string, error) {
	dir := start
	for {
		if dir == "" {
			return "", errors.New("empty start directory")
		}

		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		} else if err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// RepoPackPath returns the shared pack path for the current repository.
func RepoPackPath(start string) (string, error) {
	root, err := RepoRoot(start)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, repoPackDirName, localPackFileName), nil
}

