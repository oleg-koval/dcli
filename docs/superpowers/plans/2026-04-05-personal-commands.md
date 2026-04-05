# Spec: Personal and Shared Commands for dcli

## Objective

Let users define their own automation commands in dcli, while keeping the built-in `git` and `docker` flows intact.

The goal is twofold:

1. Personal automation should be simple enough for one user to add locally without changing Go code.
2. Useful commands should be easy to share with a team by committing them into the repo.
3. Commands should be traceable over time so users can inspect, update, and version them safely.
4. Users should have a terminal-first interface for browsing, editing, enabling, disabling, and sharing commands.
5. Users should be able to export a complete command pack for onboarding another developer.

Success looks like this:

- A user can define a namespaced command such as `dcli john docker whatever`.
- The command is discoverable through normal Cobra help and shell completion.
- The same command definition format works for both local personal commands and repo-shared commands.
- Existing built-in commands continue to work unchanged.

## Assumptions

1. This remains a Go CLI built on Cobra, not a plugin marketplace or remote code execution platform.
2. Custom commands are declarative objects with enough metadata to fully describe how they run.
3. Local personal commands live in the user config area, and shared commands live in the repository.
4. User-defined commands have higher priority than built-ins, but name collisions are resolved explicitly instead of silently shadowing behavior.
5. Repo-shared commands load automatically when the current repo contains a command pack.

## Tech Stack

- Go 1.25+
- Cobra for command trees and help text
- YAML for command definitions
- Existing config handling in `internal/config`
- Standard library execution primitives for running command steps
- A lightweight TUI library for interactive command management if the existing CLI output is not enough

## Commands

- Build: `make build`
- Test: `make test`
- Lint: `make lint`
- Dev: `go run . --help`

## Project Structure

Proposed layout:

- `cmd/`
  - Root command wiring and dynamic command registration
  - Existing built-in command groups such as `git` and `docker`
  - New helper code for loading custom command definitions
- `internal/config/`
  - Shared config parsing for local and repo-shared command definitions
- `internal/commands/`
  - Command definition parsing, validation, and execution logic
- `internal/commands/ui/`
  - Terminal UI for browsing and manipulating commands
- `~/.dcli/commands/`
  - Per-user command definitions, aliases, and history metadata
- `.dcli/commands/`
  - Repo-shared command definitions for a team or project
- `docs/CONFIGURATION.md`
  - User-facing configuration reference
- `docs/superpowers/plans/`
  - Living implementation specs and plans

## Code Style

Follow the existing Go style in this repo:

- small command files named after the feature area
- explicit error returns from `RunE`
- stdout for command output, stderr for diagnostics
- table-driven tests for behavior
- avoid shelling out through `sh -c` unless there is no safe alternative
- interactive flows should still degrade cleanly to non-interactive flags and subcommands

Example command definition model:

```go
type CommandSpec struct {
	Name        string
	Namespace   string
	Description string
	Source      string
	Enabled     bool
	Revision    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Steps       []Step
}

type Step struct {
	Type string // exec or shell
	Dir  string
	Env  map[string]string
	Args []string
	Script string
}
```

Example style for a command definition parser:

```go
func (r *Registry) Load(paths ...string) error {
	for _, path := range paths {
		defs, err := ParseFile(path)
		if err != nil {
			return fmt.Errorf("load command definitions from %s: %w", path, err)
		}
		r.Merge(defs)
	}
	return nil
}
```

## Testing Strategy

- Unit test parsing, validation, and command registration in `internal/commands`.
- Test Cobra behavior in `cmd/*_test.go` by executing commands programmatically.
- Cover success and failure paths for:
  - loading local command files
  - loading repo-shared command files
  - duplicate command names
  - invalid command definitions
  - execution of a custom command
  - listing and inspecting command metadata
  - updating a command definition and preserving its revision history
  - export and import of command packs for onboarding
  - terminal UI navigation and action selection
  - explicit conflict resolution when two definitions claim the same name
- Keep existing built-in command tests passing unchanged.

## Boundaries

- Always:
  - preserve current `git` and `docker` behavior
  - validate custom command definitions before registering them
  - keep logs and errors off stdout
  - add or update tests for any new command behavior
  - provide a non-interactive fallback for every interactive command action
  - treat secrets as runtime input only; do not persist them inside command definitions
- Ask first:
  - changing the config file format in a backward-incompatible way
  - adding a remote registry or network fetch for command packs
   - introducing arbitrary shell execution without argument validation
- Never:
  - execute untrusted remote code implicitly
  - break existing command names or help output without a migration path
  - remove current flows for Docker or Git

## Success Criteria

- A user can add a personal command definition locally and run it through dcli.
- A repo can ship shared command definitions that show up for all users on that repo.
- Command names can be nested, so a namespace like `john docker whatever` is valid.
- Custom commands appear in `dcli --help` and follow Cobra completion behavior.
- Invalid definitions fail with clear messages and do not register partial command trees.
- Users can inspect command metadata, including where a command came from and when it changed.
- Users can update or replace a saved command without editing raw code.
- Users can browse commands in a terminal UI, then run, edit, disable, or share them from the same screen.
- A developer can export a command pack for onboarding, and another developer can import it after installing dcli.
- Shared commands can be installed into a project or user profile without manually editing files.
- Conflicts never overwrite silently; the user is shown a resolution path to rename, disable, or choose the winning definition.
- Command definitions can be exported as a single JSON file suitable for onboarding and sharing.
- Existing built-in tests for `git` and `docker` still pass.

## Open Questions

1. Should the TUI be built into `dcli` directly, or exposed as a dedicated `dcli commands` subcommand with an interactive mode?
