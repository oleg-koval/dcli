package ui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oleg-koval/dcli/internal/commands"
)

type runResultMsg struct {
	path   string
	output string
	err    error
}

type exportResultMsg struct {
	path string
	err  error
}

type importResultMsg struct {
	path  string
	count int
	err   error
}

// Model is the interactive command browser.
type Model struct {
	workspace  *commands.Workspace
	reserved   [][]string
	exportPath string

	items    []commands.ResolvedCommand
	selected map[string]struct{}
	cursor   int
	width    int
	height   int
	status   string
	errMsg   string
}

// NewModel constructs a UI model from a loaded workspace.
func NewModel(workspace *commands.Workspace, reserved [][]string, exportPath string) Model {
	m := Model{
		workspace:  workspace,
		reserved:   reserved,
		exportPath: exportPath,
		selected:   make(map[string]struct{}),
	}
	m.refresh()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case " ":
			m.toggleSelected()
		case "enter", "r":
			if cmd := m.runCurrent(); cmd != nil {
				return m, cmd
			}
		case "e":
			if err := m.toggleEnabled(); err != nil {
				m.errMsg = fmt.Sprintf("could not toggle command: %v", err)
			} else {
				m.status = "updated command"
				m.refresh()
			}
		case "d":
			if err := m.deleteCurrent(); err != nil {
				m.errMsg = fmt.Sprintf("could not remove command: %v", err)
			} else {
				m.status = "removed command"
				m.refresh()
			}
		case "x":
			if cmd := m.exportSelection(); cmd != nil {
				return m, cmd
			}
		case "i":
			if cmd := m.importPack(); cmd != nil {
				return m, cmd
			}
		case "g":
			m.refresh()
			m.status = "refreshed command list"
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case runResultMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("could not run %s: %v", msg.path, msg.err)
		} else {
			output := strings.TrimSpace(msg.output)
			if output == "" {
				m.status = "command finished cleanly"
			} else {
				m.status = output
			}
		}
	case exportResultMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("could not export commands: %v", msg.err)
		} else {
			m.status = "packed up selection to " + msg.path
		}
	case importResultMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("could not import commands: %v", msg.err)
		} else {
			m.refresh()
			m.status = fmt.Sprintf("welcomed %d command(s) from %s", msg.count, msg.path)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("dcli commands ui\n")
	b.WriteString("────────────────────────────────────────\n")
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		checked := "[ ]"
		if m.isSelected(item) {
			checked = "[x]"
		}
		fmt.Fprintf(&b, "%s %s %-26s %-11s %s\n", cursor, checked, item.Command.DisplayName(), item.Status, item.Command.Description)
	}

	b.WriteString("\n")
	if len(m.items) == 0 {
		b.WriteString("No custom commands yet.\n")
		b.WriteString("Add a shortcut with `dcli commands add ...` or import a pack to get started.\n")
	} else {
		item := m.items[m.cursor]
		fmt.Fprintf(&b, "path: %s\n", item.Command.DisplayName())
		fmt.Fprintf(&b, "status: %s\n", item.Status)
		fmt.Fprintf(&b, "scope: %s\n", item.Command.Scope)
		fmt.Fprintf(&b, "source: %s\n", item.Command.Source)
		if item.ConflictReason != "" {
			fmt.Fprintf(&b, "conflict: %s\n", item.ConflictReason)
		}
		if item.Command.Description != "" {
			fmt.Fprintf(&b, "description: %s\n", item.Command.Description)
		}
		fmt.Fprintf(&b, "revision: %d\n", item.Command.Revision)
		fmt.Fprintf(&b, "steps: %d\n", len(item.Command.Steps))
	}

	if m.errMsg != "" {
		b.WriteString("\nerror: ")
		b.WriteString(m.errMsg)
		b.WriteByte('\n')
	}
	if m.status != "" {
		b.WriteString("status: ")
		b.WriteString(m.status)
		b.WriteByte('\n')
	}
	b.WriteString("keys: ↑/↓ move, space select, enter/r run, e toggle, d delete, x export, i import, g refresh, q quit\n")
	if m.exportPath != "" {
		fmt.Fprintf(&b, "export file: %s\n", m.exportPath)
	}
	return b.String()
}

func (m *Model) refresh() {
	if m.workspace == nil {
		m.items = nil
		m.cursor = 0
		return
	}
	m.items = m.workspace.ResolvedCommands(m.reserved)
	if len(m.items) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) move(delta int) {
	if len(m.items) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = len(m.items) - 1
	}
	if m.cursor >= len(m.items) {
		m.cursor = 0
	}
}

func (m Model) current() (commands.ResolvedCommand, bool) {
	if len(m.items) == 0 || m.cursor < 0 || m.cursor >= len(m.items) {
		return commands.ResolvedCommand{}, false
	}
	return m.items[m.cursor], true
}

func (m Model) keyFor(item commands.ResolvedCommand) string {
	return item.Command.Scope + "|" + item.Command.Source + "|" + item.Command.Key()
}

func (m Model) isSelected(item commands.ResolvedCommand) bool {
	_, ok := m.selected[m.keyFor(item)]
	return ok
}

func (m *Model) toggleSelected() {
	item, ok := m.current()
	if !ok {
		return
	}
	key := m.keyFor(item)
	if _, exists := m.selected[key]; exists {
		delete(m.selected, key)
		return
	}
	m.selected[key] = struct{}{}
}

func (m *Model) toggleEnabled() error {
	item, ok := m.current()
	if !ok {
		return nil
	}
	if err := m.workspace.UpdateCommand(item.Command.Path, item.Command.Scope, func(command *commands.Command) error {
		command.Enabled = !command.Enabled
		command.Revision++
		command.UpdatedAt = time.Now().UTC()
		return nil
	}); err != nil {
		return err
	}
	return saveByScope(m.workspace, item.Command.Scope)
}

func (m *Model) deleteCurrent() error {
	item, ok := m.current()
	if !ok {
		return nil
	}
	if !m.workspace.DeleteCommand(item.Command.Path, item.Command.Scope) {
		return os.ErrNotExist
	}
	return saveByScope(m.workspace, item.Command.Scope)
}

func (m Model) runCurrent() tea.Cmd {
	item, ok := m.current()
	if !ok {
		return nil
	}

	command := item.Command.Clone()
	return func() tea.Msg {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		err := commands.Execute(context.Background(), command, nil, &stdout, &stderr)
		output := stdout.String()
		if stderr.Len() > 0 {
			if output != "" {
				output += "\n"
			}
			output += stderr.String()
		}
		return runResultMsg{path: command.DisplayName(), output: output, err: err}
	}
}

func (m Model) exportSelection() tea.Cmd {
	if strings.TrimSpace(m.exportPath) == "" {
		return func() tea.Msg {
			return exportResultMsg{err: fmt.Errorf("export file is required")}
		}
	}

	selected := m.selectedCommands()
	if len(selected) == 0 {
		item, ok := m.current()
		if !ok {
			return func() tea.Msg {
				return exportResultMsg{err: fmt.Errorf("no commands available to export")}
			}
		}
		selected = []commands.Command{item.Command.Clone()}
	}

	pack := commands.Pack{Version: commands.PackVersion, Commands: selected}
	path := m.exportPath
	return func() tea.Msg {
		return exportResultMsg{path: path, err: pack.Save(path)}
	}
}

func (m Model) selectedCommands() []commands.Command {
	if len(m.selected) == 0 {
		return nil
	}
	selected := make([]commands.Command, 0, len(m.selected))
	for _, item := range m.items {
		if m.isSelected(item) {
			selected = append(selected, item.Command.Clone())
		}
	}
	return selected
}

func (m Model) importPack() tea.Cmd {
	if strings.TrimSpace(m.exportPath) == "" {
		return func() tea.Msg {
			return importResultMsg{err: fmt.Errorf("import file is required")}
		}
	}

	path := m.exportPath
	return func() tea.Msg {
		pack, err := commands.LoadPackFile(path)
		if err != nil {
			return importResultMsg{path: path, err: err}
		}
		added := 0
		for _, command := range pack.Commands {
			scope := command.Scope
			if scope != commands.ScopeShared {
				scope = commands.ScopeLocal
			}
			if scope == commands.ScopeShared && m.workspace.RepoPath == "" {
				scope = commands.ScopeLocal
			}
			if addErr := m.workspace.AddCommand(command, scope); addErr != nil {
				return importResultMsg{path: path, err: addErr}
			}
			added++
		}
		if err := m.workspace.Save(); err != nil {
			return importResultMsg{path: path, err: err}
		}
		return importResultMsg{path: path, count: added}
	}
}

func saveByScope(workspace *commands.Workspace, scope string) error {
	if scope == commands.ScopeShared {
		return workspace.SaveRepo()
	}
	return workspace.SaveLocal()
}
