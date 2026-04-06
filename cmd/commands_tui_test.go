package cmd

import (
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandsUISubcommandIsRegistered(t *testing.T) {
	command, _, err := commandsCmd.Find([]string{"ui"})
	if err != nil {
		t.Fatalf("expected commands ui to be registered: %v", err)
	}
	if command == nil {
		t.Fatal("expected ui command")
	}
}

func TestCommandsUIRunsInTestMode(t *testing.T) {
	original := commandBrowserRunner
	commandBrowserRunner = func(model tea.Model, in io.Reader, out io.Writer) error {
		return nil
	}
	defer func() { commandBrowserRunner = original }()

	t.Setenv("HOME", t.TempDir())
	repoDir := t.TempDir()
	setWorkingDirForTest(t, repoDir)

	if err := commandsUICmd.RunE(commandsUICmd, nil); err != nil {
		t.Fatalf("expected ui command to run: %v", err)
	}
}

func TestCommandsUIHelpHighlightsManagementMode(t *testing.T) {
	if !strings.Contains(commandsUICmd.Long, "management surface") {
		t.Fatalf("expected commands ui help to describe management mode, got %q", commandsUICmd.Long)
	}
}
