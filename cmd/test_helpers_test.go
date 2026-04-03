package cmd

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()

	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set %s: %v", key, err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %s: %v", key, err)
	}
}

func removeAll(t *testing.T, path string) {
	t.Helper()

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("failed to remove %s: %v", path, err)
	}
}
