package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

func TestRegisterCustomCommandsExecutesLoadedCommand(t *testing.T) {
	homeDir := t.TempDir()
	repoDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	setWorkingDirForTest(t, repoDir)
	defer func() { builtinPathsSnapshot = nil }()

	pack := commands.Pack{
		Version: commands.PackVersion,
		Commands: []commands.Command{
			{
				Path:    []string{"john", "hello"},
				Scope:   commands.ScopeLocal,
				Enabled: true,
				Steps: []commands.Step{
					{Type: commands.StepTypeShell, Script: "echo custom-command"},
				},
			},
		},
	}

	if err := os.MkdirAll(filepath.Join(homeDir, ".dcli"), 0o755); err != nil {
		t.Fatalf("failed to create pack directory: %v", err)
	}
	if err := pack.Save(filepath.Join(homeDir, ".dcli", "commands.json")); err != nil {
		t.Fatalf("failed to save pack: %v", err)
	}

	root := &cobra.Command{Use: "dcli", SilenceUsage: true, SilenceErrors: true}
	if err := registerCustomCommands(root); err != nil {
		t.Fatalf("registerCustomCommands failed: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"john", "hello"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected command to execute, got %v", err)
	}
	if !strings.Contains(stdout.String(), "custom-command") {
		t.Fatalf("expected command output, got %q", stdout.String())
	}
}

