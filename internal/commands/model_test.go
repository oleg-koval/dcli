package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandValidate(t *testing.T) {
	command := Command{
		Path: []string{"john", "docker", "restart"},
		Steps: []Step{
			{Type: StepTypeExec, Command: []string{"docker", "compose", "restart"}},
		},
	}

	if err := command.Validate(); err != nil {
		t.Fatalf("expected valid command, got %v", err)
	}
}

func TestCommandValidateRejectsInvalidStep(t *testing.T) {
	command := Command{
		Path: []string{"john", "bad"},
		Steps: []Step{{Type: "unknown"}},
	}

	if err := command.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPackRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	pack := Pack{
		Version: PackVersion,
		Commands: []Command{
			{
				Path:        []string{"john", "web", "up"},
				Description: "start local web stack",
				Scope:       ScopeLocal,
				Source:      "/tmp/commands.json",
				Enabled:     true,
				Revision:    3,
				CreatedAt:   now,
				UpdatedAt:   now,
				Steps: []Step{
					{Type: StepTypeExec, Command: []string{"docker", "compose", "up", "-d"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(pack); err != nil {
		t.Fatalf("failed to encode pack: %v", err)
	}

	var decoded Pack
	if err := json.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode pack: %v", err)
	}

	if len(decoded.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(decoded.Commands))
	}
	if decoded.Commands[0].Key() != "john web up" {
		t.Fatalf("unexpected key: %s", decoded.Commands[0].Key())
	}
}

func TestLocalPackPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	path, err := LocalPackPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(homeDir, ".dcli", "commands.json")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
}

func TestPackSaveAndLoad(t *testing.T) {
	pack := Pack{
		Version: PackVersion,
		Commands: []Command{
			{
				Path:    []string{"john", "api", "test"},
				Scope:   ScopeLocal,
				Enabled: true,
				Steps: []Step{
					{Type: StepTypeExec, Command: []string{"go", "test", "./..."}},
				},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "commands.json")
	if err := pack.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadPackFile(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(loaded.Commands))
	}
}

func TestPackSaveRejectsInvalidCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.json")
	pack := Pack{
		Version: PackVersion,
		Commands: []Command{
			{Path: []string{"broken"}},
		},
	}

	if err := pack.Save(path); err == nil {
		t.Fatal("expected save validation error")
	}
}

func TestRepoRootDetectsMissingRepo(t *testing.T) {
	root, err := RepoRoot(t.TempDir())
	if err == nil {
		t.Fatalf("expected error, got root %s", root)
	}
}

func TestWorkspaceResolveStates(t *testing.T) {
	workspace := &Workspace{
		Local: Pack{
			Version: PackVersion,
			Commands: []Command{
				{
					Path:    []string{"john", "shared"},
					Scope:   ScopeLocal,
					Enabled: true,
					Steps: []Step{{Type: StepTypeExec, Command: []string{"echo", "local"}}},
				},
				{
					Path:    []string{"john", "disabled"},
					Scope:   ScopeLocal,
					Enabled: false,
					Steps: []Step{{Type: StepTypeExec, Command: []string{"echo", "disabled"}}},
				},
			},
		},
		Repo: Pack{
			Version: PackVersion,
			Commands: []Command{
				{
					Path:    []string{"john", "shared"},
					Scope:   ScopeShared,
					Enabled: true,
					Steps: []Step{{Type: StepTypeExec, Command: []string{"echo", "repo"}}},
				},
			},
		},
	}

	resolved := workspace.ResolvedCommands(nil)
	if len(resolved) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(resolved))
	}

	var conflictCount int
	var disabledCount int
	for _, item := range resolved {
		switch item.Status {
		case StatusConflicted:
			conflictCount++
		case StatusDisabled:
			disabledCount++
		}
	}

	if conflictCount != 2 {
		t.Fatalf("expected 2 conflicts, got %d", conflictCount)
	}
	if disabledCount != 1 {
		t.Fatalf("expected 1 disabled command, got %d", disabledCount)
	}
}

func TestWorkspaceSaveLocalAndRepo(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	repoDir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create fake repo: %v", err)
	}

	workspace := &Workspace{
		Cwd:       repoDir,
		LocalPath: filepath.Join(homeDir, ".dcli", "commands.json"),
		RepoPath:  filepath.Join(repoDir, ".dcli", "commands.json"),
		Local: Pack{
			Version: PackVersion,
			Commands: []Command{
				{
					Path:    []string{"john", "local"},
					Scope:   ScopeLocal,
					Enabled: true,
					Steps: []Step{{Type: StepTypeExec, Command: []string{"echo", "local"}}},
				},
			},
		},
		Repo: Pack{
			Version: PackVersion,
			Commands: []Command{
				{
					Path:    []string{"john", "shared"},
					Scope:   ScopeShared,
					Enabled: true,
					Steps: []Step{{Type: StepTypeExec, Command: []string{"echo", "shared"}}},
				},
			},
		},
	}

	if err := workspace.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if _, err := os.Stat(workspace.LocalPath); err != nil {
		t.Fatalf("expected local pack to exist: %v", err)
	}
	if _, err := os.Stat(workspace.RepoPath); err != nil {
		t.Fatalf("expected repo pack to exist: %v", err)
	}
}
