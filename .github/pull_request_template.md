## Summary
- Provide a short, high-level overview of what this PR changes (two sentences max).
- Highlight which primary CLI flows, commands, or configuration paths the change touches.

## Problem
- Describe the user-facing or internal pain point you are addressing.
- If the change fixes a bug, summarize the incorrect behavior or confusing output.

## Solution
- Outline how you solved the problem. Mention key packages, commands, or scripts that changed.
- Call out notable updates to configuration, Docker handling, or release hooks.

## Scope
- List the affected subcommands, modules, or documentation areas (e.g., `dcli docker clean`, `internal/git`).
- Indicate whether the change is limited to docs/tests, a single command, or spans multiple helpers.

## Type of change
- [ ] `feat` (new capability)
- [ ] `fix` (bug fix)
- [ ] `docs` (documentation only)
- [ ] `test` (adds or changes tests)
- [ ] `chore` (maintenance, tooling, CI)
- [ ] `other` (describe in Summary)

## Related issues
- Link the main issue(s) this PR addresses (e.g., Closes #123). Mention other relevant discussions if any.

## Validation
- List commands you ran locally (e.g., `make test`, `go test ./...`, `bin/dcli docker clean --dry-run`).
- Include links or filenames for logs if the validation spanned multiple platforms.

## Screenshots / Demo
- Paste representative terminal output, logs, or command snippets showing the new behavior.
- If nothing visual changed, note `N/A` (screenshots are rare for CLI, but logs show the runtime effect).

## Risk and impact
- Describe deployment concerns (extra permissions, config migrations, update checks, etc.).
- Call out performance, platform, or compatibility risks for Docker Compose and Git workflows.

## Breaking changes
- Explain any backward-incompatible behavior, config file changes, or new required flags.
- If nothing breaks, write `None`.

## Documentation
- Note which docs/test plans were updated (e.g., `docs/CONFIGURATION.md`, `TESTING_README.md`).
- Mention if you updated the README, CONtributing guide, or release notes.

## Reviewer notes
- Highlight manual verification steps, simplifications, or follow-up cleanup that reviewers should know.
- Mention specific areas requiring additional scrutiny (self-updates, CI badges, config parsing).

## Checklist
- [ ] Tests pass locally (`make test`, `go test ./...`, etc.)
- [ ] No additional lint or vet warnings in `go test` output
- [ ] Documentation or docs site updated if behavior changed (e.g., `docs/`, `README.md`)
- [ ] Changelog/Release notes entry added if the change ships in a release
