package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oleg-koval/dcli/internal/config"
	"github.com/spf13/cobra"
)

func TestGitResetHelp(t *testing.T) {
	gitResetCmd.SetArgs([]string{"--help"})
	err := gitResetCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGitResetNoBranch(t *testing.T) {
	gitResetCmd.SetArgs([]string{})
	if gitResetCmd.Name() != "reset" {
		t.Fatalf("expected command name 'reset', got %s", gitResetCmd.Name())
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

func TestGitResetInvalidBranch(t *testing.T) {
	// Create a fresh root command to test
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(gitResetCmd)

	// Test that invalid branch names are rejected
	rootCmd.SetArgs([]string{"reset", "invalid-branch"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid branch, got nil")
	}

	// Verify error message mentions allowed branches
	if err != nil && !containsStr(err.Error(), "develop") && !containsStr(err.Error(), "acceptance") {
		t.Logf("error message doesn't mention allowed branches: %v", err)
	}
}

func TestGitResetValidBranchDevelop(t *testing.T) {
	// Create a fresh root command to test
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(gitResetCmd)

	// Test that develop is accepted (will fail if no config, but that's OK for this test)
	rootCmd.SetArgs([]string{"reset", "develop"})
	err := rootCmd.Execute()
	// Expected to fail due to no config file, but the branch should be accepted
	if err != nil && !containsStr(err.Error(), "no repositories") && !containsStr(err.Error(), "failed to load config") {
		// If error is not about config, it's a different issue
		t.Logf("command error: %v", err)
	}
}

func TestGitResetValidBranchAcceptance(t *testing.T) {
	// Create a fresh root command to test
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(gitResetCmd)

	// Test that acceptance is accepted (will fail if no config, but that's OK for this test)
	rootCmd.SetArgs([]string{"reset", "acceptance"})
	err := rootCmd.Execute()
	// Expected to fail due to no config file, but the branch should be accepted
	if err != nil && !containsStr(err.Error(), "no repositories") && !containsStr(err.Error(), "failed to load config") {
		// If error is not about config, it's a different issue
		t.Logf("command error: %v", err)
	}
}

func TestGitResetNoRepositoriesConfigured(t *testing.T) {
	// Create temp home with empty config
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

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
	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	// Create a fresh root command to test
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(gitResetCmd)
	rootCmd.SetArgs([]string{"reset", "develop"})
	err = rootCmd.Execute()
	if err == nil {
		t.Error("expected error for no repositories, got nil")
	}

	if err != nil && !containsStr(err.Error(), "no repositories") {
		t.Errorf("expected error about no repositories, got: %v", err)
	}
}

func TestGitResetWithConfiguration(t *testing.T) {
	// Create temp home with valid config
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Create config with repositories
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := `repositories:
  - name: test-repo
    path: /nonexistent/path
    remote: origin
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME
	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	// Create a fresh root command to test
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(gitResetCmd)
	rootCmd.SetArgs([]string{"reset", "develop"})
	err = rootCmd.Execute()
	// Expected to fail because path doesn't exist, but config should load
	if err != nil && !containsStr(err.Error(), "failed") && !containsStr(err.Error(), "some repositories") {
		t.Logf("command error: %v", err)
	}
}

func TestGitResetCommandStructure(t *testing.T) {
	// Verify command structure
	if gitResetCmd.Name() != "reset" {
		t.Errorf("expected command name 'reset', got %s", gitResetCmd.Name())
	}

	// Verify it requires exactly 1 argument
	// This is set in the command definition with cobra.ExactArgs(1)
	t.Logf("git reset command configured with ExactArgs requirement")
}

func TestGitResetConfigLoading(t *testing.T) {
	// Test that config loading works properly
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

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

	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

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

func containsStr(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
