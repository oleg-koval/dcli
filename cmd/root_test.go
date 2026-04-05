package cmd

import (
	"context"
	"strings"
	"testing"
)

type fakeStartupUpdater struct {
	called  bool
	version string
	args    []string
}

func (f *fakeStartupUpdater) Run(_ context.Context, currentVersion string, args []string) {
	f.called = true
	f.version = currentVersion
	f.args = append([]string{}, args...)
}

func TestExecuteInvokesStartupUpdater(t *testing.T) {
	originalUpdater := autoUpdateRunner
	defer func() { autoUpdateRunner = originalUpdater }()

	fake := &fakeStartupUpdater{}
	autoUpdateRunner = fake

	rootCmd.SetArgs([]string{"--help"})
	defer rootCmd.SetArgs(nil)

	Execute()

	if !fake.called {
		t.Fatal("expected startup updater to be called")
	}
	if fake.version != Version {
		t.Fatalf("expected version %q, got %q", Version, fake.version)
	}
	if len(fake.args) == 0 {
		t.Fatal("expected args to be forwarded to the startup updater")
	}
}

func TestRootHelpHighlightsExecutionFirstWorkflow(t *testing.T) {
	if !strings.Contains(rootCmd.Long, "execution-first") {
		t.Fatalf("expected root help to mention execution-first workflow, got %q", rootCmd.Long)
	}
	if !strings.Contains(rootCmd.Long, "dcli commands ui") {
		t.Fatalf("expected root help to mention command management UI, got %q", rootCmd.Long)
	}
}
