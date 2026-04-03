//go:build gofuzz
// +build gofuzz

package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// FuzzConfigYAML fuzzes YAML configuration parsing and unmarshaling
func FuzzConfigYAML(f *testing.F) {
	testcases := []string{
		"repositories: []",
		"repositories:\n  - path: /tmp\n    name: test",
		"",
		"invalid: yaml: structure:",
		"repositories:\n  - {}",
		"repositories:\n  - path: /path\n    name: repo\n    remote: origin",
		"repositories:\n  - path: /path\n    name: repo\n    remote: origin\n    invalid_field: value",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Empty input short-circuits (same as actual behavior)
		if input == "" {
			return
		}

		// Exercise the YAML parsing logic used by Load()
		var cfg Config
		err := yaml.Unmarshal([]byte(input), &cfg)

		// The fuzzer validates parsing doesn't panic on arbitrary input
		// err can be nil (valid YAML) or non-nil (invalid YAML)
		// Both cases should be handled gracefully without panics
		_ = err

		// Validate parsed structure if no error
		if err == nil && cfg.Repositories != nil {
			// Exercise field access to ensure no panics on malformed structures
			for _, repo := range cfg.Repositories {
				_ = repo.Name
				_ = repo.Path
				_ = repo.Remote
			}
		}
	})
}
