package cmd

import (
	"testing"
)

func TestGitResetHelp(t *testing.T) {
	gitResetCmd.SetArgs([]string{"--help"})
	err := gitResetCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGitResetNoBranch(t *testing.T) {
	gitResetCmd.SetArgs([]string{})
	if gitResetCmd.Name() != "reset" {
		t.Fatalf("expected command name 'reset', got %s", gitResetCmd.Name())
	}
}
