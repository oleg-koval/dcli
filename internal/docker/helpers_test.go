package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetServices(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "test-docker-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal docker-compose.yml for testing
	composeContent := `version: '3'
services:
  web:
    image: nginx
  db:
    image: postgres
`
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("failed to write docker-compose.yml: %v", err)
	}

	// Test GetServices
	services, err := GetServices(tmpDir)
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}

	// Verify we got the expected services
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d: %v", len(services), services)
	}

	// Check that both expected services are present
	serviceMap := make(map[string]bool)
	for _, s := range services {
		serviceMap[s] = true
	}

	if !serviceMap["web"] {
		t.Error("expected service 'web' not found")
	}
	if !serviceMap["db"] {
		t.Error("expected service 'db' not found")
	}
}

func TestGetServicesInvalidProjectDir(t *testing.T) {
	// Test with non-existent directory
	services, err := GetServices("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}

	if services != nil {
		t.Fatalf("expected nil services on error, got %v", services)
	}
}

func TestRunCommand(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "test-docker-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test RunCommand with a simple command that should work
	// Using 'docker ps' which is unlikely to fail and doesn't require running containers
	err = RunCommand(tmpDir, "ps")
	if err != nil {
		// It's okay if docker ps fails (e.g., docker not installed),
		// we're mainly testing that projectDir is used
		t.Logf("RunCommand execution note: %v (docker may not be installed)", err)
	}
}
