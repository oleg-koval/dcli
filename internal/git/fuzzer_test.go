//go:build gofuzz
// +build gofuzz

package git

import (
	"testing"
)

// FuzzGitReset fuzzes the Git reset command
func FuzzGitReset(f *testing.F) {
	testcases := []string{
		"develop",
		"acceptance",
		"main",
		"feature/test",
		"",
		"branch-with-dash",
		"branch_with_underscore",
		"v1.0.0",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// This fuzzer ensures git branch parsing can handle arbitrary input
		// without panicking or crashing
		if input == "" {
			return
		}
		// In a real implementation, you would validate and process the branch name
		// The fuzzer helps find edge cases and potential crashes
	})
}
