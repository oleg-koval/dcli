package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Workspace represents local and repo-shared command packs together.
type Workspace struct {
	Cwd       string
	LocalPath string
	RepoPath  string
	Local     Pack
	Repo      Pack
}

// LoadWorkspace loads the current user's command packs.
func LoadWorkspace(cwd string) (*Workspace, error) {
	localPath, err := LocalPackPath()
	if err != nil {
		return nil, fmt.Errorf("resolve local pack path: %w", err)
	}

	localPack, err := LoadPackFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("load local pack: %w", err)
	}

	repoPath, repoPack, err := loadRepoPack(cwd)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Cwd:       cwd,
		LocalPath: localPath,
		RepoPath:  repoPath,
		Local:     localPack,
		Repo:      repoPack,
	}, nil
}

// LoadPackFile loads a command pack from disk.
func LoadPackFile(path string) (Pack, error) {
	path = filepath.Clean(path)
	//nolint:gosec // G304: path is resolved by callers to the local or repo pack file only.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Pack{Version: PackVersion}, nil
		}
		return Pack{}, fmt.Errorf("read %s: %w", path, err)
	}

	var pack Pack
	if err := json.Unmarshal(data, &pack); err != nil {
		return Pack{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if pack.Version == 0 {
		pack.Version = PackVersion
	}
	return pack, nil
}

// Save writes the pack to disk using a stable JSON encoding.
func (p Pack) Save(path string) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create pack directory: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal pack: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write pack: %w", err)
	}
	return nil
}

// EnsureVersion sets the pack version if needed.
func (p *Pack) EnsureVersion() {
	if p.Version == 0 {
		p.Version = PackVersion
	}
}

// AddCommand appends a command to the pack after normalizing it.
func (p *Pack) AddCommand(command Command, scope string, source string) error {
	command.Normalize(scope, source)
	if err := command.Validate(); err != nil {
		return err
	}
	p.Commands = append(p.Commands, command.Clone())
	p.EnsureVersion()
	return nil
}

// UpdateCommand updates an existing command by path.
func (p *Pack) UpdateCommand(path []string, fn func(*Command) error) error {
	index := p.indexOf(path)
	if index < 0 {
		return os.ErrNotExist
	}

	command := p.Commands[index].Clone()
	if err := fn(&command); err != nil {
		return err
	}
	command.UpdatedAt = command.UpdatedAt.UTC()
	if err := command.Validate(); err != nil {
		return err
	}
	p.Commands[index] = command
	return nil
}

// DeleteCommand removes a command from the pack by path.
func (p *Pack) DeleteCommand(path []string) bool {
	index := p.indexOf(path)
	if index < 0 {
		return false
	}
	p.Commands = append(p.Commands[:index], p.Commands[index+1:]...)
	return true
}

// FindCommand returns a command by path.
func (p Pack) FindCommand(path []string) (Command, bool) {
	index := p.indexOf(path)
	if index < 0 {
		return Command{}, false
	}
	return p.Commands[index].Clone(), true
}

func (p Pack) indexOf(path []string) int {
	for i, command := range p.Commands {
		if equalPath(command.Path, path) {
			return i
		}
	}
	return -1
}

func equalPath(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func loadRepoPack(cwd string) (string, Pack, error) {
	repoRoot, err := RepoRoot(cwd)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", Pack{Version: PackVersion}, nil
		}
		return "", Pack{}, fmt.Errorf("resolve repository root: %w", err)
	}

	repoPath := filepath.Join(repoRoot, repoPackDirName, localPackFileName)
	pack, err := LoadPackFile(repoPath)
	if err != nil {
		return "", Pack{}, err
	}
	return repoPath, pack, nil
}

// ResolvedCommands returns all commands with state annotations.
func (w *Workspace) ResolvedCommands(reserved [][]string) []ResolvedCommand {
	commands := make([]ResolvedCommand, 0, len(w.Local.Commands)+len(w.Repo.Commands))
	appendPack := func(pack Pack) {
		for _, command := range pack.Commands {
			commands = append(commands, ResolvedCommand{Command: command.Clone()})
		}
	}

	appendPack(w.Repo)
	appendPack(w.Local)

	resolveStates(commands, reserved)
	SortCommands(commands)
	return commands
}

// ActiveCommands returns commands that can be registered with Cobra.
func (w *Workspace) ActiveCommands(reserved [][]string) []ResolvedCommand {
	commands := w.ResolvedCommands(reserved)
	active := commands[:0]
	for _, command := range commands {
		if command.Status == StatusActive {
			active = append(active, command)
		}
	}
	return active
}

// SaveLocal persists the local pack.
func (w *Workspace) SaveLocal() error {
	w.Local.EnsureVersion()
	return w.Local.Save(w.LocalPath)
}

// SaveRepo persists the repo pack when a repository exists.
func (w *Workspace) SaveRepo() error {
	if w.RepoPath == "" {
		return os.ErrNotExist
	}
	w.Repo.EnsureVersion()
	return w.Repo.Save(w.RepoPath)
}

// Save persists both packs where available.
func (w *Workspace) Save() error {
	if err := w.SaveLocal(); err != nil {
		return err
	}
	if w.RepoPath != "" {
		if err := w.SaveRepo(); err != nil {
			return err
		}
	}
	return nil
}

// AddCommand stores a command in the requested scope.
func (w *Workspace) AddCommand(command Command, scope string) error {
	source := w.LocalPath
	target := &w.Local
	if scope == ScopeShared {
		if w.RepoPath == "" {
			return os.ErrNotExist
		}
		source = w.RepoPath
		target = &w.Repo
	}
	if err := target.AddCommand(command, scope, source); err != nil {
		return err
	}
	return nil
}

// DeleteCommand removes a command from the requested scope.
func (w *Workspace) DeleteCommand(path []string, scope string) bool {
	if scope == ScopeShared {
		return w.Repo.DeleteCommand(path)
	}
	return w.Local.DeleteCommand(path)
}

// UpdateCommand updates a command in the requested scope.
func (w *Workspace) UpdateCommand(path []string, scope string, fn func(*Command) error) error {
	if scope == ScopeShared {
		return w.Repo.UpdateCommand(path, fn)
	}
	return w.Local.UpdateCommand(path, fn)
}

func resolveStates(commands []ResolvedCommand, reserved [][]string) {
	conflicted := make(map[int]string)

	for i, command := range commands {
		if !command.Command.Enabled {
			commands[i].Status = StatusDisabled
		} else {
			commands[i].Status = StatusActive
		}
	}

	for i := range commands {
		if commands[i].Status != StatusActive {
			continue
		}
		for j := i + 1; j < len(commands); j++ {
			if commands[j].Status != StatusActive {
				continue
			}
			if pathPrefixConflict(commands[i].Command.Path, commands[j].Command.Path) {
				conflicted[i] = commands[j].Command.Key()
				conflicted[j] = commands[i].Command.Key()
			}
		}
		for _, path := range reserved {
			if pathPrefixConflict(commands[i].Command.Path, path) {
				conflicted[i] = "built-in command"
				break
			}
		}
	}

	for index, reason := range conflicted {
		commands[index].Status = StatusConflicted
		commands[index].ConflictReason = reason
	}
}

// PackSummary returns a deterministic presentation order for commands.
func PackSummary(commands []Command) []Command {
	sorted := make([]Command, len(commands))
	copy(sorted, commands)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Key() < sorted[j].Key()
	})
	return sorted
}
