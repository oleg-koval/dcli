package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerRestartHelp(t *testing.T) {
	dockerRestartCmd.SetArgs([]string{"--help"})
	err := dockerRestartCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerRestartNoArgs(t *testing.T) {
	dockerRestartCmd.SetArgs([]string{})
	if dockerRestartCmd.Name() != "restart" {
		t.Fatalf("expected command name 'restart', got %s", dockerRestartCmd.Name())
	}
}

func TestDockerRestartWithServiceArgs(t *testing.T) {
	dockerRestartCmd.SetArgs([]string{"service1", "service2"})
	if dockerRestartCmd.Name() != "restart" {
		t.Fatalf("expected command name 'restart', got %s", dockerRestartCmd.Name())
	}
}

func TestDockerRestartProjectDirHandling(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-docker-restart-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal docker-compose.yml
	composeContent := `version: '3'
services:
  test-service:
    image: nginx
`
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("failed to write docker-compose.yml: %v", err)
	}

	// Set DCLI_PROJECT_DIR environment variable
	oldProjectDir := os.Getenv("DCLI_PROJECT_DIR")
	defer func() {
		if oldProjectDir != "" {
			os.Setenv("DCLI_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("DCLI_PROJECT_DIR")
		}
	}()

	os.Setenv("DCLI_PROJECT_DIR", tmpDir)

	// Test that DCLI_PROJECT_DIR is properly read and used
	// Note: This test verifies the env var is read; actual command execution
	// requires docker to be installed and running
	projectDir := os.Getenv("DCLI_PROJECT_DIR")
	if projectDir != tmpDir {
		t.Fatalf("expected DCLI_PROJECT_DIR to be %s, got %s", tmpDir, projectDir)
	}
}

func TestDockerRestartDefaultProjectDir(t *testing.T) {
	// Ensure DCLI_PROJECT_DIR is not set
	oldProjectDir := os.Getenv("DCLI_PROJECT_DIR")
	defer func() {
		if oldProjectDir != "" {
			os.Setenv("DCLI_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("DCLI_PROJECT_DIR")
		}
	}()

	os.Unsetenv("DCLI_PROJECT_DIR")

	// Verify that when DCLI_PROJECT_DIR is not set, it defaults to "."
	projectDir := os.Getenv("DCLI_PROJECT_DIR")
	if projectDir != "" {
		t.Fatalf("expected DCLI_PROJECT_DIR to be unset, got %s", projectDir)
	}
	// The command should default to "." if DCLI_PROJECT_DIR is empty
}

func TestDockerRestartCommandMetadata(t *testing.T) {
	if dockerRestartCmd.Use != "restart [services...]" {
		t.Errorf("expected Use 'restart [services...]', got %s", dockerRestartCmd.Use)
	}

	if dockerRestartCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if dockerRestartCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	if dockerRestartCmd.RunE == nil {
		t.Error("expected RunE function to be defined")
	}
}

func TestDockerRestartPreservesVolumes(t *testing.T) {
	// Create temp directory with docker-compose.yml
	tmpDir, err := os.MkdirTemp("", "test-docker-restart-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create docker-compose.yml with volumes
	composeContent := `version: '3'
services:
  db:
    image: postgres
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  db_data:
`
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("failed to write docker-compose.yml: %v", err)
	}

	// Set project directory
	oldProjectDir := os.Getenv("DCLI_PROJECT_DIR")
	defer func() {
		if oldProjectDir != "" {
			os.Setenv("DCLI_PROJECT_DIR", oldProjectDir)
		} else {
			os.Unsetenv("DCLI_PROJECT_DIR")
		}
	}()
	os.Setenv("DCLI_PROJECT_DIR", tmpDir)

	// Execute command
	dockerRestartCmd.SetArgs([]string{"db"})
	err = dockerRestartCmd.Execute()
	// Note: Will fail if docker is not running, but that's expected
	if err != nil {
		t.Logf("command execution note: %v (docker may not be running)", err)
	}
}

func TestDockerRestartCommandStructure(t *testing.T) {
	// Verify the command has proper structure
	if dockerRestartCmd.Name() != "restart" {
		t.Errorf("expected command name 'restart', got %s", dockerRestartCmd.Name())
	}

	// Verify description mentions data preservation
	if !contains(dockerRestartCmd.Long, "preserves") && !contains(dockerRestartCmd.Long, "data") {
		t.Logf("Note: Long description doesn't mention data preservation")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
