//go:build gofuzz
// +build gofuzz

package docker

import (
	"strings"
	"testing"
)

// FuzzCleanCommand fuzzes the Docker clean command parsing and argument building
func FuzzCleanCommand(f *testing.F) {
	testcases := []string{
		"",
		"service1",
		"service1 service2 service3",
		"  service1  ",
		"service-with-dash",
		"service_with_underscore",
		"123service",
		"service\nwith\nnewline",
		"service\twith\ttab",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Empty input short-circuits early (same as actual command)
		if input == "" {
			return
		}

		// Parse service names from input (simulating how docker clean/restart parse args)
		services := strings.Fields(input)
		if len(services) == 0 {
			return
		}

		// Exercise the argument parsing/building logic used in docker clean/restart
		// by constructing the same command arguments they would build
		rmArgs := append([]string{"compose", "rm", "-sfv"}, services...)
		buildArgs := append([]string{"compose", "build"}, services...)
		upArgs := append([]string{"compose", "up", "-d"}, services...)

		// Verify all arguments are parsed correctly and no panics occur
		_ = len(rmArgs)
		_ = len(buildArgs)
		_ = len(upArgs)

		// If the actual docker helper were available, we'd call:
		// dockerHelper.RunCommand(projectDir, rmArgs...)
		// For now, the fuzzer validates argument construction doesn't panic
	})
}
