package docker

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// GetServices retrieves a list of services from docker-compose config.
// Optional profiles are passed as --profile flags to docker compose.
func GetServices(projectDir string, profiles ...string) ([]string, error) {
	args := slices.Concat(composePrefix(profiles), []string{"config", "--services"})

	cmd := exec.Command("docker", args...) // #nosec G204 -- args passed to exec.Command without shell interpolation
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	services := strings.Fields(string(output))
	return services, nil
}

// RunCommand executes a Docker command with the given arguments in the specified project directory.
// Stdout and stderr are piped to the terminal so the user sees build/restart progress.
func RunCommand(projectDir string, args ...string) error {
	cmd := exec.Command("docker", args...) // #nosec G204 -- args are passed directly to docker without shell expansion
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker command failed: %w", err)
	}
	return nil
}

// GetContainers retrieves a list of running Docker containers
func GetContainers() ([]string, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}") // #nosec G204 -- fixed command, no shell interpolation
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

// composePrefix returns ["compose", "--profile", p1, "--profile", p2, ...] for the given profiles.
func composePrefix(profiles []string) []string {
	args := []string{"compose"}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	return args
}

// BuildCleanCommandArgs builds docker compose arguments for clean operation.
// Profiles are injected as --profile flags right after "compose".
func BuildCleanCommandArgs(services []string, profiles ...string) (rmArgs, buildArgs, upArgs []string) {
	prefix := composePrefix(profiles)
	rmArgs = slices.Concat(prefix, []string{"rm", "-sfv"}, services)
	buildArgs = slices.Concat(prefix, []string{"build"}, services)
	upArgs = slices.Concat(prefix, []string{"up", "-d"}, services)
	return rmArgs, buildArgs, upArgs
}

// BuildRestartCommandArgs builds docker compose arguments for restart operation.
// Profiles are injected as --profile flags right after "compose".
func BuildRestartCommandArgs(services []string, profiles ...string) (stopArgs, upArgs []string) {
	prefix := composePrefix(profiles)
	stopArgs = slices.Concat(prefix, []string{"stop"}, services)
	upArgs = slices.Concat(prefix, []string{"up", "-d"}, services)
	return stopArgs, upArgs
}
