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

		// Parse service names from input (same parsing used by docker clean/restart commands)
		services := strings.Fields(input)
		if len(services) == 0 {
			return
		}

		// Exercise the argument parsing and building logic used by docker clean command
		// This routes the fuzzed input into the real code paths used by the CLI
		rmArgs, buildArgs, upArgs := BuildCleanCommandArgs(services)

		// Verify all arguments are correctly built and accessible without panics
		// The fuzzer exercises the argument construction and parsing logic
		if len(rmArgs) == 0 || len(buildArgs) == 0 || len(upArgs) == 0 {
			t.Errorf("expected non-empty args, got rm:%d build:%d up:%d", len(rmArgs), len(buildArgs), len(upArgs))
		}

		// Also test the restart command args to exercise that path
		stopArgs, restartUpArgs := BuildRestartCommandArgs(services)
		if len(stopArgs) == 0 || len(restartUpArgs) == 0 {
			t.Errorf("expected non-empty restart args, got stop:%d up:%d", len(stopArgs), len(restartUpArgs))
		}
	})
}
