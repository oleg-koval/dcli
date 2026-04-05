package commands

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	PackVersion = 1

	ScopeLocal  = "local"
	ScopeShared = "shared"

	StatusActive     = "active"
	StatusDisabled   = "disabled"
	StatusConflicted = "conflicted"

	StepTypeExec  = "exec"
	StepTypeShell = "shell"
)

// Pack stores a complete command collection.
type Pack struct {
	Version  int       `json:"version"`
	Commands []Command `json:"commands"`
}

// Command describes one automation command.
type Command struct {
	Path        []string  `json:"path"`
	Description string    `json:"description,omitempty"`
	Scope       string    `json:"scope"`
	Source      string    `json:"source,omitempty"`
	Enabled     bool      `json:"enabled"`
	Revision    int       `json:"revision"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Steps       []Step    `json:"steps"`
}

// Step describes one command action.
type Step struct {
	Type    string            `json:"type"`
	Dir     string            `json:"dir,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Command []string          `json:"command,omitempty"`
	Script  string            `json:"script,omitempty"`
}

// ResolvedCommand annotates a command with its runtime state.
type ResolvedCommand struct {
	Command        Command `json:"command"`
	Status         string  `json:"status"`
	ConflictReason string  `json:"conflict_reason,omitempty"`
}

// Validate checks a pack for malformed commands.
func (p Pack) Validate() error {
	for i, command := range p.Commands {
		if err := command.Validate(); err != nil {
			return fmt.Errorf("command %d: %w", i, err)
		}
	}
	return nil
}

// Normalize fills default metadata.
func (c *Command) Normalize(scope string, source string) {
	c.Scope = scope
	c.Source = source
	if c.Revision <= 0 {
		c.Revision = 1
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}
}

// Key returns the canonical command path.
func (c Command) Key() string {
	return strings.Join(c.Path, " ")
}

// DisplayName returns a user-friendly command path.
func (c Command) DisplayName() string {
	return c.Key()
}

// Validate checks a single command definition.
func (c Command) Validate() error {
	if len(c.Path) == 0 {
		return errors.New("path is required")
	}
	for i, segment := range c.Path {
		if strings.TrimSpace(segment) == "" {
			return fmt.Errorf("path segment %d is empty", i)
		}
		if segment != strings.TrimSpace(segment) {
			return fmt.Errorf("path segment %d has surrounding spaces", i)
		}
	}
	if len(c.Steps) == 0 {
		return errors.New("at least one step is required")
	}
	for i, step := range c.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks one step definition.
func (s Step) Validate() error {
	switch s.Type {
	case StepTypeExec:
		if len(s.Command) == 0 {
			return errors.New("exec step requires command")
		}
	case StepTypeShell:
		if strings.TrimSpace(s.Script) == "" {
			return errors.New("shell step requires script")
		}
	default:
		return fmt.Errorf("unsupported step type %q", s.Type)
	}

	for key := range s.Env {
		if strings.TrimSpace(key) == "" {
			return errors.New("environment variable name cannot be empty")
		}
	}
	return nil
}

// Clone returns a deep copy of the command.
func (c Command) Clone() Command {
	clone := c
	clone.Path = append([]string(nil), c.Path...)
	clone.Steps = make([]Step, len(c.Steps))
	for i := range c.Steps {
		clone.Steps[i] = c.Steps[i].Clone()
	}
	return clone
}

// Clone returns a deep copy of a step.
func (s Step) Clone() Step {
	clone := s
	clone.Command = append([]string(nil), s.Command...)
	if len(s.Env) > 0 {
		clone.Env = make(map[string]string, len(s.Env))
		for k, v := range s.Env {
			clone.Env[k] = v
		}
	}
	return clone
}

// SortCommands sorts commands by path for stable presentation.
func SortCommands(commands []ResolvedCommand) {
	sort.SliceStable(commands, func(i, j int) bool {
		return commands[i].Command.Key() < commands[j].Command.Key()
	})
}

func pathPrefixConflict(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

