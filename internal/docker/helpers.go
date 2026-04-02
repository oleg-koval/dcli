package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetServices retrieves a list of running Docker services
func GetServices() ([]string, error) {
	cmd := exec.Command("docker", "service", "ls", "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker services: %w", err)
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty strings
	var filtered []string
	for _, s := range services {
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	return filtered, nil
}

// RunCommand executes a Docker command with the given arguments
func RunCommand(args ...string) error {
	cmd := exec.Command("docker", args...)
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
	var filtered []string
	for _, c := range containers {
		if c != "" {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}
