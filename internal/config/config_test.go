package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".dcli", "config.yaml")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}
}

func TestLoadConfigFileNotExist(t *testing.T) {
	// Create temp home directory
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Override HOME environment variable
	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error when file doesn't exist, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Repositories) != 0 {
		t.Errorf("expected empty repositories, got %d", len(cfg.Repositories))
	}
}

func TestLoadConfigValid(t *testing.T) {
	// Create temp home directory
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Create .dcli directory
	dcliDir := filepath.Join(tmpHome, ".dcli")
	if err := os.MkdirAll(dcliDir, 0755); err != nil {
		t.Fatalf("failed to create .dcli dir: %v", err)
	}

	// Write config file
	configFile := filepath.Join(dcliDir, "config.yaml")
	configContent := `repositories:
  - name: backend
    path: /Users/user/projects/backend
    remote: origin
  - name: frontend
    path: /Users/user/projects/frontend
    remote: origin
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Override HOME environment variable
	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Repositories) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Name != "backend" {
		t.Errorf("expected first repo name 'backend', got %s", cfg.Repositories[0].Name)
	}

	if cfg.Repositories[0].Path != "/Users/user/projects/backend" {
		t.Errorf("expected path /Users/user/projects/backend, got %s", cfg.Repositories[0].Path)
	}

	if cfg.Repositories[1].Name != "frontend" {
		t.Errorf("expected second repo name 'frontend', got %s", cfg.Repositories[1].Name)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
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
  - name: backend
    path: /Users/user/projects/backend
  broken yaml here [[[]`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
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

	_, err = Load()
	if err == nil {
		t.Fatal("expected error loading invalid YAML, got nil")
	}
}

func TestLoadConfigEmptyFile(t *testing.T) {
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
	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty config file: %v", err)
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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Empty YAML results in nil repositories, which is valid
	if cfg.Repositories != nil && len(cfg.Repositories) > 0 {
		t.Errorf("expected empty or nil repositories, got %d", len(cfg.Repositories))
	}
}

func TestSaveConfigCreateDir(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	cfg := &Config{
		Repositories: []Repository{
			{
				Name:   "test",
				Path:   "/test/path",
				Remote: "origin",
			},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	configFile := filepath.Join(tmpHome, ".dcli", "config.yaml")
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestSaveConfigWrite(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	cfg := &Config{
		Repositories: []Repository{
			{
				Name:   "backend",
				Path:   "/Users/user/backend",
				Remote: "origin",
			},
			{
				Name:   "frontend",
				Path:   "/Users/user/frontend",
				Remote: "upstream",
			},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load config back and verify
	configFile := filepath.Join(tmpHome, ".dcli", "config.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("config file is empty")
	}

	// Verify it contains expected content
	content := string(data)
	if !contains(content, "backend") {
		t.Error("saved config doesn't contain 'backend'")
	}
	if !contains(content, "/Users/user/backend") {
		t.Error("saved config doesn't contain backend path")
	}
	if !contains(content, "frontend") {
		t.Error("saved config doesn't contain 'frontend'")
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	os.Setenv("HOME", tmpHome)

	// Create original config
	originalCfg := &Config{
		Repositories: []Repository{
			{
				Name:   "repo1",
				Path:   "/path/1",
				Remote: "origin",
			},
			{
				Name:   "repo2",
				Path:   "/path/2",
				Remote: "upstream",
			},
		},
	}

	// Save it
	if err := originalCfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load it back
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if len(loadedCfg.Repositories) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loadedCfg.Repositories))
	}

	for i, repo := range loadedCfg.Repositories {
		if repo.Name != originalCfg.Repositories[i].Name {
			t.Errorf("repo %d: expected name %s, got %s", i, originalCfg.Repositories[i].Name, repo.Name)
		}
		if repo.Path != originalCfg.Repositories[i].Path {
			t.Errorf("repo %d: expected path %s, got %s", i, originalCfg.Repositories[i].Path, repo.Path)
		}
		if repo.Remote != originalCfg.Repositories[i].Remote {
			t.Errorf("repo %d: expected remote %s, got %s", i, originalCfg.Repositories[i].Remote, repo.Remote)
		}
	}
}

func TestRepositoryStruct(t *testing.T) {
	repo := Repository{
		Name:   "test-repo",
		Path:   "/path/to/repo",
		Remote: "origin",
	}

	if repo.Name != "test-repo" {
		t.Errorf("expected name 'test-repo', got %s", repo.Name)
	}
	if repo.Path != "/path/to/repo" {
		t.Errorf("expected path '/path/to/repo', got %s", repo.Path)
	}
	if repo.Remote != "origin" {
		t.Errorf("expected remote 'origin', got %s", repo.Remote)
	}
}

func TestConfigEmpty(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{},
	}

	if len(cfg.Repositories) != 0 {
		t.Errorf("expected empty repositories, got %d", len(cfg.Repositories))
	}
}

func TestConfigNilRepositories(t *testing.T) {
	cfg := &Config{}

	if cfg.Repositories != nil && len(cfg.Repositories) > 0 {
		t.Errorf("expected nil or empty repositories, got %v", cfg.Repositories)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
