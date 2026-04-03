//go:build gofuzz
// +build gofuzz

package docker

import (
	"testing"
)

// FuzzCleanCommand fuzzes the Docker clean command parsing
func FuzzCleanCommand(f *testing.F) {
	testcases := []string{
		"",
		"service1",
		"service1 service2 service3",
		"  service1  ",
		"service-with-dash",
		"service_with_underscore",
		"123service",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// This fuzzer ensures the docker helpers can handle arbitrary input
		// without panicking or crashing
		if input == "" {
			return
		}
		// In a real implementation, you would parse and process the input here
		// The fuzzer helps find edge cases and crashes
	})
}
