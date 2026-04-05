package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oleg-koval/dcli/internal/commands"
	"github.com/spf13/cobra"
)

func TestCommandsAddPersistsLocalPack(t *testing.T) {
	homeDir := t.TempDir()
	repoDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	setWorkingDirForTest(t, repoDir)

	root := &cobra.Command{Use: "dcli", SilenceUsage: true, SilenceErrors: true}
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.AddCommand(commandsCmd)
	root.SetArgs([]string{"commands", "add", "john", "deploy", "--", "echo", "deploying"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected add command to succeed, got %v", err)
	}

	workspace, err := commands.LoadWorkspace(repoDir)
	if err != nil {
		t.Fatalf("load workspace failed: %v", err)
	}

	if len(workspace.Local.Commands) != 1 {
		t.Fatalf("expected 1 local command, got %d", len(workspace.Local.Commands))
	}
	if got := workspace.Local.Commands[0].Key(); got != "john deploy" {
		t.Fatalf("unexpected command key: %s", got)
	}
	if _, err := os.Stat(filepath.Join(homeDir, ".dcli", "commands.json")); err != nil {
		t.Fatalf("expected local pack file to exist: %v", err)
	}
	if !strings.Contains(stdout.String(), "Added john deploy to your local pack.") {
		t.Fatalf("expected friendly add confirmation, got %q", stdout.String())
	}
}

func TestCommandsHelpHighlightsManagementAndExecutionSplit(t *testing.T) {
	if !strings.Contains(commandsCmd.Long, "Execution stays separate from management") {
		t.Fatalf("expected commands help to describe the execution/management split, got %q", commandsCmd.Long)
	}
}
