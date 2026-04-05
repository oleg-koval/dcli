# Implementation Plan: Personal and Shared Commands for dcli

## Overview

Build a command-pack system for dcli that lets users create local personal shortcuts, load repo-shared commands automatically, inspect and manage them in a terminal UI, and export/import a full pack for onboarding.

The implementation should preserve the existing built-in `git` and `docker` flows, make command conflicts explicit, and keep secrets out of persisted command definitions.

## Architecture Decisions

- Commands are stored as declarative JSON/YAML objects with full metadata, not as generated Go code.
- Repo-shared command packs load automatically from the current repository; local command packs load from the user profile and layer on top.
- User-defined commands must not silently override each other; conflicts require explicit resolution.
- Built-in commands remain first-class Cobra commands, while custom commands are registered dynamically after config load.
- The CLI should expose `dcli commands` as the management entrypoint, with interactive TUI support and non-interactive flags for scripting.
- Export/import should use a single pack file that contains command definitions plus metadata for onboarding and sharing.

## Task List

### Phase 1: Command Model and Loading

- [ ] Task 1: Define the command-pack data model and persistence layer
  - **Description:** Add the types and serialization logic for command definitions, revisions, metadata, enabled state, and step execution details. Add load/save helpers for local and repo-shared packs.
  - **Acceptance criteria:**
    - [ ] A command pack can be serialized and deserialized without losing metadata.
    - [ ] Command steps can represent `exec` and shell-based execution.
    - [ ] Secrets are not stored in the persisted command definition model.
  - **Verification:**
    - [ ] Unit tests cover round-trip serialization.
    - [ ] Unit tests cover validation failures for malformed definitions.
  - **Dependencies:** None
  - **Files likely touched:**
    - `internal/commands/*.go`
    - `internal/commands/*_test.go`
    - `internal/config/config.go`
  - **Estimated scope:** Medium

- [ ] Task 2: Implement discovery, precedence, and conflict resolution
  - **Description:** Load built-ins, repo-shared packs, and local packs in the chosen order. Detect name collisions and surface them as explicit conflicts instead of silently shadowing.
  - **Acceptance criteria:**
    - [ ] Repo-shared packs load automatically when present in the current repo.
    - [ ] Local commands are loaded after repo-shared commands.
    - [ ] Conflicts are detected and reported with enough context to resolve them.
  - **Verification:**
    - [ ] Command registry tests verify precedence and conflict handling.
    - [ ] Existing built-in command execution remains unchanged.
  - **Dependencies:** Task 1
  - **Files likely touched:**
    - `cmd/root.go`
    - `cmd/*.go`
    - `internal/commands/*.go`
    - `cmd/*_test.go`
  - **Estimated scope:** Medium

### Checkpoint: Foundation

- [ ] Tests pass for pack parsing, loading, and precedence
- [ ] Built-in commands still resolve and execute
- [ ] Conflicts are visible and do not silently overwrite

### Phase 2: CLI Management Commands

- [ ] Task 3: Add `dcli commands` management subcommands
  - **Description:** Add the non-interactive CLI surface for listing, showing, adding, editing, disabling/enabling, deleting, exporting, and importing commands.
  - **Acceptance criteria:**
    - [ ] Users can manage commands without entering the TUI.
    - [ ] The subcommands support machine-friendly output where applicable.
    - [ ] Export writes one pack file; import reads the same format.
  - **Verification:**
    - [ ] Cobra tests exercise each management path.
    - [ ] Help output includes the new command family.
  - **Dependencies:** Tasks 1-2
  - **Files likely touched:**
    - `cmd/commands.go`
    - `cmd/commands_add.go`
    - `cmd/commands_list.go`
    - `cmd/commands_show.go`
    - `cmd/commands_edit.go`
    - `cmd/commands_import.go`
    - `cmd/commands_export.go`
    - `cmd/*_test.go`
  - **Estimated scope:** Large

- [ ] Task 4: Add command execution plumbing
  - **Description:** Wire custom command execution into dcli so loaded commands can be run like normal subcommands, including argument passing, env injection, and working-directory handling.
  - **Acceptance criteria:**
    - [ ] Custom commands execute from the CLI with args forwarded correctly.
    - [ ] Step-level env and working-directory settings are honored.
    - [ ] Failure paths return useful errors without corrupting output streams.
  - **Verification:**
    - [ ] Unit tests cover successful and failing execution.
    - [ ] Integration-style command tests cover argument forwarding and env handling.
  - **Dependencies:** Tasks 1-3
  - **Files likely touched:**
    - `internal/commands/*.go`
    - `cmd/*.go`
    - `cmd/*_test.go`
  - **Estimated scope:** Medium

### Checkpoint: Core CLI

- [ ] Commands can be created, listed, inspected, exported, imported, and executed
- [ ] `dcli commands` works in non-interactive mode
- [ ] Pack import/export is stable and round-trips cleanly

### Phase 3: TUI and Onboarding

- [ ] Task 5: Build the terminal UI for command management
  - **Description:** Create the interactive screen for browsing command packs, selecting commands, and performing actions such as run, edit, disable, share, export, and import.
  - **Acceptance criteria:**
    - [ ] Users can navigate command packs with keyboard-only interaction.
    - [ ] Users can trigger the same management actions available in the CLI.
    - [ ] The UI shows command source, revision, enabled state, and conflict status.
  - **Verification:**
    - [ ] UI tests cover core navigation and action selection.
    - [ ] Manual smoke test confirms the TUI opens and renders loaded commands.
  - **Dependencies:** Tasks 1-4
  - **Files likely touched:**
    - `internal/commands/ui/*.go`
    - `cmd/commands.go`
    - `cmd/*_test.go`
  - **Estimated scope:** Large

- [ ] Task 6: Add docs, onboarding flow, and migration notes
  - **Description:** Document the command pack format, how a teammate installs shared commands, and how users migrate any existing config to the new system.
  - **Acceptance criteria:**
    - [ ] README and configuration docs explain local vs repo-shared packs.
    - [ ] Onboarding steps are clear enough for a new developer to follow without help.
    - [ ] Migration guidance exists if the old config format needs to remain supported.
  - **Verification:**
    - [ ] Documentation review matches actual CLI behavior.
    - [ ] Examples in docs are executable or close to executable.
  - **Dependencies:** Tasks 1-5
  - **Files likely touched:**
    - `README.md`
    - `docs/CONFIGURATION.md`
    - `docs/INSTALL.md`
    - `docs/superpowers/plans/2026-04-05-personal-commands.md`
  - **Estimated scope:** Small

### Checkpoint: Complete

- [ ] All tests pass
- [ ] Custom commands work locally and from repo-shared packs
- [ ] TUI and CLI both manage the same underlying data
- [ ] Documentation matches implemented behavior

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Command name conflicts create confusing behavior | High | Resolve conflicts explicitly and show source metadata in list/help/TUI views |
| TUI scope grows too large | High | Keep the TUI as a thin front-end over the same command service used by the CLI |
| Export/import format drifts from runtime model | Medium | Make pack serialization the single source of truth and test round-trips |
| Secrets are accidentally persisted | High | Keep secret values out of the data model and only accept them at runtime |
| Dynamic registration breaks built-in commands | High | Load custom commands after built-ins and keep built-in names protected by default |

## Open Questions

- Should the TUI be available by default when running `dcli commands`, or behind an explicit `--interactive` flag?
- Should export/import support single-command packs as well as full workspace packs from day one?
- Do we want a strict JSON-only pack format, or JSON with optional YAML import compatibility?

