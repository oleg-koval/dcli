package cmd

import (
	"os"
	"testing"
)

func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set %s: %v", key, err)
	}

	t.Cleanup(func() {
		var err error
		if hadValue {
			err = os.Setenv(key, oldValue)
		} else {
			err = os.Unsetenv(key)
		}
		if err != nil {
			t.Errorf("failed to restore %s: %v", key, err)
		}
	})
}

func cleanupDirForTest(t *testing.T, path string) {
	t.Helper()

	t.Cleanup(func() {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("cleanup failed for %s: %v", path, err)
		}
	})
}

func setWorkingDirForTest(t *testing.T, path string) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to read working directory: %v", err)
	}
	if err := os.Chdir(path); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})
}
