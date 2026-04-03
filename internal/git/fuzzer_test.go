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
		"   ",
		"branch-with-dash",
		"branch_with_underscore",
		"v1.0.0",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// This fuzzer exercises the branch validation logic with arbitrary input
		// ensuring no panics and proper error handling of invalid branch names.
		_ = ValidateBranchTarget(input)
	})
}
