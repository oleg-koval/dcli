package docker

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// GetServices retrieves a list of services from docker-compose config.
// GetServices runs `docker compose config --services` in the specified project directory and returns the configured service names.
// If profiles are provided they are passed as `--profile <name>` flags to the compose command.
// The command output is split on whitespace to produce the returned service name slice.
// If the docker command fails, an error is returned.
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
// RunCommand runs the specified docker subcommand with args inside projectDir and streams stdout and stderr to the current terminal.
// It returns an error wrapping the underlying execution failure if the docker process exits with a non-zero status.
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

// GetContainers returns the names of running Docker containers.
// It runs `docker ps --format "{{.Names}}"`, splits the output by newline,
// filters out empty names, and returns the resulting slice. An error is
// returned if the `docker` command fails.
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

// composePrefix returns an argument slice starting with "compose" followed by "--profile" and each profile name in the order provided.
func composePrefix(profiles []string) []string {
	args := []string{"compose"}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	return args
}

// BuildCleanCommandArgs builds docker compose arguments for clean operation.
//   - upArgs: arguments to start the services detached (`compose up -d ...`)
func BuildCleanCommandArgs(services []string, profiles ...string) (rmArgs, buildArgs, upArgs []string) {
	prefix := composePrefix(profiles)
	rmArgs = slices.Concat(prefix, []string{"rm", "-sfv"}, services)
	buildArgs = slices.Concat(prefix, []string{"build"}, services)
	upArgs = slices.Concat(prefix, []string{"up", "-d"}, services)
	return rmArgs, buildArgs, upArgs
}

// BuildRestartCommandArgs builds docker compose arguments for restart operation.
// BuildRestartCommandArgs builds Docker Compose argument lists to stop then start the specified services,
// injecting each provided profile as a `--profile <name>` flag immediately after `compose`.
// The first returned slice is the arguments for `docker compose stop <services>`; the second is for `docker compose up -d <services>`.
func BuildRestartCommandArgs(services []string, profiles ...string) (stopArgs, upArgs []string) {
	prefix := composePrefix(profiles)
	stopArgs = slices.Concat(prefix, []string{"stop"}, services)
	upArgs = slices.Concat(prefix, []string{"up", "-d"}, services)
	return stopArgs, upArgs
}
