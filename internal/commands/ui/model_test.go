package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oleg-koval/dcli/internal/commands"
)

func TestModelNavigationAndSelection(t *testing.T) {
	workspace := testWorkspace(t)
	model := NewModel(workspace, nil, "")

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = next.(Model)
	if model.cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", model.cursor)
	}

	next, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = next.(Model)
	if len(model.selected) != 1 {
		t.Fatalf("expected one selected command, got %d", len(model.selected))
	}
}

func TestModelToggleEnable(t *testing.T) {
	workspace := testWorkspace(t)
	model := NewModel(workspace, nil, "")

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	_ = next.(Model)

	reloaded, err := commands.LoadPackFile(workspace.LocalPath)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	command, ok := reloaded.FindCommand([]string{"john", "local"})
	if !ok {
		t.Fatal("expected reloaded command to exist")
	}
	if command.Enabled {
		t.Fatal("expected command to be disabled after toggle")
	}
}

func TestModelExportSelection(t *testing.T) {
	workspace := testWorkspace(t)
	exportPath := filepath.Join(t.TempDir(), "pack.json")
	model := NewModel(workspace, nil, exportPath)

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = next.(Model)
	cmd := model.exportSelection()
	msg := cmd()
	result, ok := msg.(exportResultMsg)
	if !ok {
		t.Fatalf("unexpected export message type %T", msg)
	}
	if result.err != nil {
		t.Fatalf("export failed: %v", result.err)
	}
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("expected export file to exist: %v", err)
	}
}

func TestModelImportPack(t *testing.T) {
	workspace := testWorkspace(t)
	exportPath := filepath.Join(t.TempDir(), "pack.json")
	model := NewModel(workspace, nil, exportPath)

	// Export current selection (fallbacks to current command when nothing selected).
	exportMsg := model.exportSelection()().(exportResultMsg)
	if exportMsg.err != nil {
		t.Fatalf("export failed: %v", exportMsg.err)
	}

	workspace.Local.Commands = nil
	if err := workspace.SaveLocal(); err != nil {
		t.Fatalf("save local failed: %v", err)
	}

	importMsg := model.importPack()().(importResultMsg)
	if importMsg.err != nil {
		t.Fatalf("import failed: %v", importMsg.err)
	}
	if importMsg.count == 0 {
		t.Fatal("expected imported commands")
	}
}

func TestModelViewShowsEmptyStateGuidance(t *testing.T) {
	workspace := testWorkspace(t)
	workspace.Local.Commands = nil
	if err := workspace.SaveLocal(); err != nil {
		t.Fatalf("save local failed: %v", err)
	}

	model := NewModel(workspace, nil, "")
	view := model.View()
	if !strings.Contains(view, "No custom commands yet.") {
		t.Fatalf("expected empty state guidance, got %q", view)
	}
	if !strings.Contains(view, "Add a shortcut with `dcli commands add ...`") {
		t.Fatalf("expected setup hint in empty state, got %q", view)
	}
}

func testWorkspace(t *testing.T) *commands.Workspace {
	t.Helper()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	localPath := filepath.Join(homeDir, ".dcli", "commands.json")
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	workspace := &commands.Workspace{
		LocalPath: localPath,
		Local: commands.Pack{
			Version: commands.PackVersion,
			Commands: []commands.Command{
				{
					Path:    []string{"john", "local"},
					Scope:   commands.ScopeLocal,
					Enabled: true,
					Steps:   []commands.Step{{Type: commands.StepTypeExec, Command: []string{"echo", "local"}}},
				},
				{
					Path:    []string{"john", "second"},
					Scope:   commands.ScopeLocal,
					Enabled: true,
					Steps:   []commands.Step{{Type: commands.StepTypeExec, Command: []string{"echo", "second"}}},
				},
			},
		},
	}

	if err := workspace.Local.Save(localPath); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	return workspace
}
