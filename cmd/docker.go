package cmd

import (
	"fmt"
	"os"

	"github.com/oleg-koval/dcli/internal/docker"
	"github.com/spf13/cobra"
)

// DockerHelper defines the interface for Docker operations
type DockerHelper interface {
	GetServices(projectDir string, profiles ...string) ([]string, error)
	RunCommand(projectDir string, args ...string) error
	GetContainers() ([]string, error)
}

// Global helper - will be overridden in tests
var dockerHelper DockerHelper = &defaultDockerHelper{}

type defaultDockerHelper struct{}

func (d *defaultDockerHelper) GetServices(projectDir string, profiles ...string) ([]string, error) {
	return docker.GetServices(projectDir, profiles...)
}

func (d *defaultDockerHelper) RunCommand(projectDir string, args ...string) error {
	return docker.RunCommand(projectDir, args...)
}

func (d *defaultDockerHelper) GetContainers() ([]string, error) {
	return docker.GetContainers()
}

// dockerProfiles holds --profile flag values for docker compose commands
var dockerProfiles []string

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker Compose management commands",
}

func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.PersistentFlags().StringSliceVar(&dockerProfiles, "profile", nil, "Docker Compose profile(s) to activate (can be specified multiple times)")
	dockerCmd.AddCommand(dockerCleanCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
}

// resolveProjectDir returns DCLI_PROJECT_DIR if set, otherwise os.Getwd().
func resolveProjectDir() (string, error) {
	if dir := os.Getenv("DCLI_PROJECT_DIR"); dir != "" {
		return dir, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to determine project directory: %w", err)
	}
	return dir, nil
}
