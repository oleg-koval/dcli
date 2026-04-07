package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDockerCleanWithValidServices(t *testing.T) {
	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string, profiles ...string) ([]string, error) {
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
		GetServicesFn: func(projectDir string, profiles ...string) ([]string, error) {
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
		GetServicesFn: func(projectDir string, profiles ...string) ([]string, error) {
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
		GetServicesFn: func(projectDir string, profiles ...string) ([]string, error) {
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

func TestDockerCleanWithProfile(t *testing.T) {
	// Set profiles directly (mirrors what --profile flag does)
	dockerProfiles = []string{"all_services"}
	t.Cleanup(func() { dockerProfiles = nil })

	mockHelper := &MockDockerHelper{
		GetServicesFn: func(projectDir string, profiles ...string) ([]string, error) {
			if len(profiles) != 1 || profiles[0] != "all_services" {
				t.Errorf("expected profiles [all_services], got %v", profiles)
			}
			return []string{"web", "api"}, nil
		},
		RunCommandFn: func(projectDir string, args ...string) error {
			// Verify --profile appears in compose args
			found := false
			for i, arg := range args {
				if arg == "--profile" && i+1 < len(args) && args[i+1] == "all_services" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected --profile all_services in args, got %v", args)
			}
			return nil
		},
	}
	setDockerHelper(mockHelper)
	defer resetDockerHelper()

	// Use a fresh command to avoid stale Cobra state from other tests
	localCleanCmd := &cobra.Command{
		Use:  dockerCleanCmd.Use,
		RunE: dockerCleanCmd.RunE,
	}
	rootCmd := &cobra.Command{}
	dockerCmd := &cobra.Command{Use: "docker"}
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(localCleanCmd)

	rootCmd.SetArgs([]string{"docker", "clean"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// GetServices should have been called with the profile
	if len(mockHelper.Calls.GetServices) == 0 {
		t.Fatal("expected GetServices to be called")
	}
	if len(mockHelper.Calls.GetServices[0].Profiles) != 1 || mockHelper.Calls.GetServices[0].Profiles[0] != "all_services" {
		t.Errorf("expected GetServices called with profiles [all_services], got %v", mockHelper.Calls.GetServices[0].Profiles)
	}

	// All 3 RunCommand calls (rm, build, up) should contain --profile
	if len(mockHelper.Calls.RunCommand) != 3 {
		t.Fatalf("expected 3 RunCommand calls, got %d", len(mockHelper.Calls.RunCommand))
	}
}
