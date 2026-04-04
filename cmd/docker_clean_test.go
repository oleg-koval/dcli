package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDockerCleanWithValidServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web", "db"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify projectDir is passed correctly
			if projectDir == "" {
				t.Error("expected non-empty projectDir")
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean", "web", "db"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify docker helper was called
	if len(mockHelper.Calls.RunCommand) == 0 {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerCleanWithNoServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web", "db", "cache"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify projectDir is passed correctly when GetServices retrieves all services
			if projectDir == "" {
				t.Error("expected non-empty projectDir")
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify GetServices was called to get all services
	if len(mockHelper.Calls.GetServices) == 0 {
		t.Error("expected GetServices to be called when no services specified")
	}
}

func TestDockerCleanRunCommandCalled(t *testing.T) {
	runCommandCalled := false
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			runCommandCalled = true
			// Verify projectDir is passed correctly
			if projectDir == "" {
				t.Error("expected non-empty projectDir")
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean", "web"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !runCommandCalled {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerCleanCommandMetadata(t *testing.T) {
	if dockerCleanCmd.Use != "clean [services...]" {
		t.Errorf("expected Use 'clean [services...]', got %s", dockerCleanCmd.Use)
	}

	if dockerCleanCmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if dockerCleanCmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	if dockerCleanCmd.RunE == nil {
		t.Error("expected RunE function to be defined")
	}
}

func TestDockerCleanProjectDirFromEnv(t *testing.T) {
	setEnvForTest(t, "DCLI_PROJECT_DIR", "/test/path")

	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			if projectDir != "/test/path" {
				t.Errorf("expected projectDir '/test/path', got %s", projectDir)
			}
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify projectDir is correctly passed from DCLI_PROJECT_DIR env var
			if projectDir != "/test/path" {
				t.Errorf("RunCommand: expected projectDir '/test/path', got %s", projectDir)
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean", "web"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerCleanHelp(t *testing.T) {
	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
