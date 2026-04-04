package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDockerRestartWithValidServices(t *testing.T) {
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
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "web", "db"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify docker helper was called
	if len(mockHelper.Calls.RunCommand) == 0 {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerRestartWithNoServices(t *testing.T) {
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
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify GetServices was called to get all services
	if len(mockHelper.Calls.GetServices) == 0 {
		t.Error("expected GetServices to be called when no services specified")
	}
}

func TestDockerRestartRunCommandCalled(t *testing.T) {
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
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "web"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !runCommandCalled {
		t.Error("expected RunCommand to be called")
	}
}

func TestDockerRestartPreservesVolumes(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"db"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify projectDir is passed correctly
			if projectDir == "" {
				t.Error("expected non-empty projectDir")
			}
			// Verify that the command does NOT include "rm" which would remove volumes
			// Restart uses "compose stop" and "compose up -d" which preserves volumes
			if len(args) > 0 {
				if args[0] == "compose" && len(args) > 1 {
					if args[1] == "rm" {
						t.Error("expected restart to not use 'rm' command (would delete volumes)")
					}
				}
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "db"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify that stop and up commands were called (not rm/rebuild)
	if len(mockHelper.Calls.RunCommand) != 2 {
		t.Errorf("expected exactly 2 RunCommand calls (stop + up), got %d", len(mockHelper.Calls.RunCommand))
	}

	// Verify first call is stop
	if len(mockHelper.Calls.RunCommand) > 0 {
		firstCall := mockHelper.Calls.RunCommand[0]
		if len(firstCall.Args) < 2 || firstCall.Args[1] != "stop" {
			t.Error("expected first RunCommand call to be 'docker compose stop'")
		}
	}

	// Verify second call is up
	if len(mockHelper.Calls.RunCommand) > 1 {
		secondCall := mockHelper.Calls.RunCommand[1]
		if len(secondCall.Args) < 2 || secondCall.Args[1] != "up" {
			t.Error("expected second RunCommand call to be 'docker compose up -d'")
		}
	}
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

func TestDockerRestartProjectDirFromEnv(t *testing.T) {
	setEnvForTest(t, "DCLI_PROJECT_DIR", "/test/path")

	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			if projectDir != "/test/path" {
				t.Errorf("expected projectDir '/test/path', got %s", projectDir)
			}
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			if projectDir != "/test/path" {
				t.Errorf("expected projectDir '/test/path', got %s", projectDir)
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "web"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerRestartHelp(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDockerRestartMultipleServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string) ([]string, error) {
			return []string{"web", "db", "cache"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify projectDir is passed correctly for multiple services
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

	// Create a fresh copy of the restart command to avoid state pollution
	localRestartCmd := &cobra.Command{
		Use:   dockerRestartCmd.Use,
		Short: dockerRestartCmd.Short,
		Long:  dockerRestartCmd.Long,
		RunE:  dockerRestartCmd.RunE,
	}
	dockerCmd.AddCommand(localRestartCmd)

	rootCmd.SetArgs([]string{"docker", "restart", "web", "db", "cache"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify that all services are included in the command calls
	if len(mockHelper.Calls.RunCommand) < 2 {
		t.Errorf("expected at least 2 RunCommand calls, got %d", len(mockHelper.Calls.RunCommand))
	}

	// Verify all expected services are passed to RunCommand
	expectedServices := map[string]bool{"web": false, "db": false, "cache": false}
	for _, call := range mockHelper.Calls.RunCommand {
		for _, arg := range call.Args {
			if _, exists := expectedServices[arg]; exists {
				expectedServices[arg] = true
			}
		}
	}
	for service, found := range expectedServices {
		if !found {
			t.Errorf("expected service %q to be passed to RunCommand", service)
		}
	}
}
