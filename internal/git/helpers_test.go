package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Helper function to initialize a git repository
func initGitRepo(t *testing.T, path string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to set git user.email: %v", err)
	}
}

// Helper function to create an initial commit
func createInitialCommit(t *testing.T, path string) {
	// Create a dummy file
	dummyFile := filepath.Join(path, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// Add and commit
	cmd := exec.Command("git", "add", "dummy.txt")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}
}

// Helper function to create a branch
func createBranch(t *testing.T, path string, branch string) {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch %s: %v", branch, err)
	}

	// Create a commit on the branch
	dummyFile := filepath.Join(path, "branch_file.txt")
	if err := os.WriteFile(dummyFile, []byte("branch content"), 0644); err != nil {
		t.Fatalf("failed to create file on branch: %v", err)
	}

	cmd = exec.Command("git", "add", "branch_file.txt")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add on branch: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "branch commit")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit on branch: %v", err)
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test non-git directory
	if IsGitRepo(tmpDir) {
		t.Error("expected false for non-git directory")
	}

	// Initialize git repo
	initGitRepo(t, tmpDir)

	// Test git directory
	if !IsGitRepo(tmpDir) {
		t.Error("expected true for git directory")
	}
}

func TestIsGitRepoNonExistent(t *testing.T) {
	nonExistent := "/nonexistent/path/to/repo"
	if IsGitRepo(nonExistent) {
		t.Error("expected false for non-existent path")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize and create initial commit
	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Test getting current branch (should be main or master)
	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	if branch == "" {
		t.Error("expected non-empty branch name")
	}

	// Verify it's either main or master
	if branch != "main" && branch != "master" {
		t.Logf("current branch: %s (expected main or master)", branch)
	}
}

func TestCheckoutBranch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize and create initial commit
	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Create a branch
	createBranch(t, tmpDir, "develop")

	// Checkout back to main/master
	mainBranch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	// Checkout to develop
	if err := CheckoutBranch(tmpDir, "develop"); err != nil {
		t.Fatalf("CheckoutBranch failed: %v", err)
	}

	// Verify we're on develop
	currentBranch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}

	if currentBranch != "develop" {
		t.Errorf("expected branch 'develop', got '%s'", currentBranch)
	}

	// Checkout back to main/master
	if err := CheckoutBranch(tmpDir, mainBranch); err != nil {
		t.Fatalf("CheckoutBranch back to %s failed: %v", mainBranch, err)
	}
}

func TestCheckoutBranchNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Try to checkout non-existent branch
	if err := CheckoutBranch(tmpDir, "nonexistent"); err == nil {
		t.Error("expected error for non-existent branch, got nil")
	}
}

func TestResetHard(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Create a branch and make changes
	createBranch(t, tmpDir, "feature")

	// Go back to main/master
	mainBranch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("failed to get main branch: %v", err)
	}

	if err := CheckoutBranch(tmpDir, mainBranch); err != nil {
		t.Fatalf("failed to checkout main: %v", err)
	}

	// Modify a file
	dummyFile := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("modified content"), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Reset hard to HEAD
	if err := ResetHard(tmpDir, "HEAD"); err != nil {
		t.Fatalf("ResetHard failed: %v", err)
	}

	// Verify file content is restored
	data, err := os.ReadFile(dummyFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != "test" {
		t.Errorf("expected file content 'test', got '%s'", string(data))
	}
}

func TestResetHardToCommit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Make a second commit
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("second commit"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	cmd := exec.Command("git", "add", "file2.txt")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add file2: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "second commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Reset to initial commit
	if err := ResetHard(tmpDir, "HEAD~1"); err != nil {
		t.Fatalf("ResetHard to HEAD~1 failed: %v", err)
	}

	// Verify file2 doesn't exist
	if _, err := os.Stat(file2); err == nil {
		t.Error("expected file2 to be deleted after reset")
	}
}

func TestResetHardInvalidTarget(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Try to reset to non-existent target
	if err := ResetHard(tmpDir, "nonexistent"); err == nil {
		t.Error("expected error for invalid reset target, got nil")
	}
}

func TestGetRemoteURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Add a remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add remote: %v", err)
	}

	// Get remote URL
	url, err := GetRemoteURL(tmpDir)
	if err != nil {
		t.Fatalf("GetRemoteURL failed: %v", err)
	}

	expectedURL := "https://github.com/test/repo.git"
	if url != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, url)
	}
}

func TestGetRemoteURLNoRemote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Try to get remote URL when no remote is configured
	_, err = GetRemoteURL(tmpDir)
	if err == nil {
		t.Error("expected error for no remote, got nil")
	}
}

func TestFetchOriginNoRemote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Try to fetch when no remote is configured
	// This should fail because there's no origin remote
	err = FetchOrigin(tmpDir)
	if err == nil {
		t.Error("expected error for fetch with no remote, got nil")
	}
}

func TestFetchOriginBadNetwork(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Add a non-existent remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://nonexistent.example.com/repo.git")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to add remote: %v", err)
	}

	// Try to fetch - should fail due to network/non-existent repo
	err = FetchOrigin(tmpDir)
	if err == nil {
		t.Error("expected error for fetch from non-existent remote, got nil")
	}
}

func TestGitRepoWithSubdirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-git-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	initGitRepo(t, tmpDir)
	createInitialCommit(t, tmpDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Subdirectory doesn't have .git folder directly, so IsGitRepo returns false
	if IsGitRepo(subDir) {
		t.Error("expected subdirectory without .git to return false")
	}

	// But GetCurrentBranch should still work because it uses git -C flag
	// which finds the repository in parent directories
	branch, err := GetCurrentBranch(subDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch from subdirectory failed: %v", err)
	}

	if branch == "" {
		t.Error("expected non-empty branch name from subdirectory")
	}
}
