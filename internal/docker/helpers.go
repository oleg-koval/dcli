package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetServices retrieves a list of services from docker-compose config
func GetServices(projectDir string) ([]string, error) {
	cmd := exec.Command("docker", "compose", "config", "--services")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := strings.Fields(string(output))
	return services, nil
}

// RunCommand executes a Docker command with the given arguments in the specified project directory
func RunCommand(projectDir string, args ...string) error {
	cmd := exec.Command("docker", args...)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker command failed: %w", err)
	}
	return nil
}

// GetContainers retrieves a list of running Docker containers
func GetContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty strings
	filtered := []string{}
	for _, c := range containers {
		if c != "" {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

// BuildCleanCommandArgs builds docker compose arguments for clean operation
// Takes a list of service names and returns the command args for rm, build, and up operations
func BuildCleanCommandArgs(services []string) (rmArgs, buildArgs, upArgs []string) {
	rmArgs = append([]string{"compose", "rm", "-sfv"}, services...)
	buildArgs = append([]string{"compose", "build"}, services...)
	upArgs = append([]string{"compose", "up", "-d"}, services...)
	return rmArgs, buildArgs, upArgs
}

// BuildRestartCommandArgs builds docker compose arguments for restart operation
// Takes a list of service names and returns the command args for stop and up operations
func BuildRestartCommandArgs(services []string) (stopArgs, upArgs []string) {
	stopArgs = append([]string{"compose", "stop"}, services...)
	upArgs = append([]string{"compose", "up", "-d"}, services...)
	return stopArgs, upArgs
}
