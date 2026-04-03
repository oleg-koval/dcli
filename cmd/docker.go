package cmd

import (
	"github.com/oleg-koval/dcli/internal/docker"
	"github.com/spf13/cobra"
)

// DockerHelper defines the interface for Docker operations
type DockerHelper interface {
	GetServices(projectDir string) ([]string, error)
	RunCommand(projectDir string, args ...string) error
	GetContainers() ([]string, error)
}

// Global helper - will be overridden in tests
var dockerHelper DockerHelper = &defaultDockerHelper{}

type defaultDockerHelper struct{}

func (d *defaultDockerHelper) GetServices(projectDir string) ([]string, error) {
	return docker.GetServices(projectDir)
}

func (d *defaultDockerHelper) RunCommand(projectDir string, args ...string) error {
	return docker.RunCommand(projectDir, args...)
}

func (d *defaultDockerHelper) GetContainers() ([]string, error) {
	return docker.GetContainers()
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker Compose management commands",
}

func init() {
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerCleanCmd)
	dockerCmd.AddCommand(dockerRestartCmd)
}
