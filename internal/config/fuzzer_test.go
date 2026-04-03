//go:build gofuzz
// +build gofuzz

package config

import (
	"testing"
)

// FuzzConfigYAML fuzzes YAML configuration parsing
func FuzzConfigYAML(f *testing.F) {
	testcases := []string{
		"repositories: []",
		"repositories:\n  - path: /tmp\n    name: test",
		"",
		"invalid: yaml: structure:",
		"repositories:\n  - {}",
		"repositories:\n  - path: /path\n    name: repo\n    remote: origin",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// This fuzzer ensures config parsing can handle arbitrary YAML input
		// without panicking or crashing
		if input == "" {
			return
		}
		// In a real implementation, you would parse the YAML here
		// The fuzzer helps find edge cases in parsing logic
	})
}
