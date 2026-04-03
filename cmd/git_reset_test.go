package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/oleg-koval/dcli/internal/config"
	"github.com/spf13/cobra"
)

func TestGitResetValidBranch(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	setEnvForTest(t, "HOME", tmpHome)

	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify git helper was called
	if len(mockHelper.Calls.FetchOrigin) == 0 {
		t.Error("expected FetchOrigin to be called")
	}
	if len(mockHelper.Calls.CheckoutBranch) == 0 {
		t.Error("expected CheckoutBranch to be called")
	}
	if len(mockHelper.Calls.ResetHard) == 0 {
		t.Error("expected ResetHard to be called")
	}
}

func TestGitResetAcceptanceBranch(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	setEnvForTest(t, "HOME", tmpHome)

	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "acceptance"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify git operations were called
	if len(mockHelper.Calls.FetchOrigin) == 0 {
		t.Error("expected FetchOrigin to be called")
	}
	if len(mockHelper.Calls.CheckoutBranch) == 0 {
		t.Error("expected CheckoutBranch to be called")
	}
	if len(mockHelper.Calls.ResetHard) == 0 {
		t.Error("expected ResetHard to be called")
	}
}

func TestGitResetInvalidBranch(t *testing.T) {
	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "invalid-branch"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid branch, got nil")
	}

	// Verify error message mentions allowed branches
	if err != nil {
		errMsg := err.Error()
		if errMsg != "branch must be 'develop' or 'acceptance', got 'invalid-branch'" {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}

func TestGitResetFetchOrigin(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	setEnvForTest(t, "HOME", tmpHome)

	fetchOriginCalled := false
	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			fetchOriginCalled = true
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !fetchOriginCalled {
		t.Error("expected FetchOrigin to be called")
	}
}

func TestGitResetCheckoutBranch(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	var checkoutBranch string
	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			checkoutBranch = branch
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "acceptance"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if checkoutBranch != "acceptance" {
		t.Errorf("expected checkout branch 'acceptance', got '%s'", checkoutBranch)
	}
}

func TestGitResetHardReset(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	var resetTarget string
	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			// Capture the branch argument passed to ResetHard
			resetTarget = branch
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify ResetHard was called with the remote ref (origin/develop)
	expectedRef := "origin/develop"
	if resetTarget != expectedRef {
		t.Errorf("expected ResetHard to be called with remote ref %q, got %q", expectedRef, resetTarget)
	}
}

func TestGitResetConfigLoading(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := `repositories:
  - name: repo1
    path: /path/1
    remote: origin
  - name: repo2
    path: /path/2
    remote: upstream
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	setEnvForTest(t, "HOME", tmpHome)

	// Load config directly to verify it works
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Name != "repo1" {
		t.Errorf("expected first repo name 'repo1', got %s", cfg.Repositories[0].Name)
	}

	if cfg.Repositories[1].Remote != "upstream" {
		t.Errorf("expected second repo remote 'upstream', got %s", cfg.Repositories[1].Remote)
	}
}

func TestGitResetEmptyConfig(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create empty config
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("repositories: []"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("expected error for no repositories, got nil")
	}

	if err != nil {
		errMsg := err.Error()
		if errMsg != "no repositories configured in ~/.dcli/config.yaml" {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}

func TestGitResetMultipleRepos(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with multiple repos
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directories
	repo1Path := filepath.Join(tmpHome, "repo1")
	repo2Path := filepath.Join(tmpHome, "repo2")
	repo3Path := filepath.Join(tmpHome, "repo3")
	for _, p := range []string{repo1Path, repo2Path, repo3Path} {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: repo1
    path: %s
    remote: origin
  - name: repo2
    path: %s
    remote: origin
  - name: repo3
    path: %s
    remote: origin
`, repo1Path, repo2Path, repo3Path)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify git operations were called for each repo
	if len(mockHelper.Calls.FetchOrigin) != 3 {
		t.Errorf("expected FetchOrigin to be called 3 times, got %d", len(mockHelper.Calls.FetchOrigin))
	}
	if len(mockHelper.Calls.CheckoutBranch) != 3 {
		t.Errorf("expected CheckoutBranch to be called 3 times, got %d", len(mockHelper.Calls.CheckoutBranch))
	}
	if len(mockHelper.Calls.ResetHard) != 3 {
		t.Errorf("expected ResetHard to be called 3 times, got %d", len(mockHelper.Calls.ResetHard))
	}
}

func TestGitResetCommandMetadata(t *testing.T) {
	if gitResetCmd.Use != "reset [develop|acceptance]" {
		t.Errorf("expected Use 'reset [develop|acceptance]', got %s", gitResetCmd.Use)
	}

	if gitResetCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if gitResetCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	if gitResetCmd.RunE == nil {
		t.Error("expected RunE function to be defined")
	}
}

func TestGitResetWithHelp(t *testing.T) {
	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	gitCmdLocal.AddCommand(gitResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGitResetFetchOriginBeforeCheckout(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	callOrder := []string{}
	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			callOrder = append(callOrder, "fetch")
			return nil
		},
		CheckoutBranchFn: func(path, branch string) error {
			callOrder = append(callOrder, "checkout")
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			callOrder = append(callOrder, "reset")
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	localResetCmd := &cobra.Command{
		Use:   gitResetCmd.Use,
		Short: gitResetCmd.Short,
		Long:  gitResetCmd.Long,
		Args:  gitResetCmd.Args,
		RunE:  gitResetCmd.RunE,
	}
	gitCmdLocal.AddCommand(localResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify call order: fetch -> checkout -> reset
	if len(callOrder) < 3 {
		t.Errorf("expected at least 3 calls, got %d", len(callOrder))
	}
	if len(callOrder) >= 3 {
		if callOrder[0] != "fetch" {
			t.Errorf("expected first call to be 'fetch', got '%s'", callOrder[0])
		}
		if callOrder[1] != "checkout" {
			t.Errorf("expected second call to be 'checkout', got '%s'", callOrder[1])
		}
		if callOrder[2] != "reset" {
			t.Errorf("expected third call to be 'reset', got '%s'", callOrder[2])
		}
	}
}

func TestGitResetFetchOriginError(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	cleanupDirForTest(t, tmpHome)

	// Create config with test repo
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Create test repo directory
	repoPath := filepath.Join(tmpHome, "test-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := fmt.Sprintf(`repositories:
  - name: test-repo
    path: %s
    remote: origin
`, repoPath)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	setEnvForTest(t, "HOME", tmpHome)

	mockHelper := &MockGitHelper{
		IsGitRepoFn: func(path string) bool {
			return true
		},
		FetchOriginFn: func(path string) error {
			return fmt.Errorf("fetch failed")
		},
		CheckoutBranchFn: func(path, branch string) error {
			return nil
		},
		ResetHardFn: func(path, branch string) error {
			return nil
		},
	}
	setGitHelper(mockHelper)
	defer resetGitHelper()

	rootCmd := &cobra.Command{}
	gitCmdLocal := &cobra.Command{Use: "git"}
	rootCmd.AddCommand(gitCmdLocal)
	localResetCmd := &cobra.Command{
		Use:   gitResetCmd.Use,
		Short: gitResetCmd.Short,
		Long:  gitResetCmd.Long,
		Args:  gitResetCmd.Args,
		RunE:  gitResetCmd.RunE,
	}
	gitCmdLocal.AddCommand(localResetCmd)

	rootCmd.SetArgs([]string{"git", "reset", "develop"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("expected error when fetch fails, got nil")
	}

	if err != nil && err.Error() != "some repositories failed to reset" {
		t.Fatalf("expected 'some repositories failed to reset', got %v", err)
	}
}
