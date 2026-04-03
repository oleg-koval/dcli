# Repository Guidelines

## Project Structure & Module Organization
`dcli` is a small Go CLI built around Cobra. Entry points live in `main.go` and `cmd/`, where each subcommand has its own file such as `cmd/docker_clean.go` or `cmd/git_reset.go`. Reusable logic lives under `internal/` (`internal/docker`, `internal/git`, `internal/config`). Project docs are in `docs/`, and shared repo metadata lives at the root (`README.md`, `SECURITY.md`, `CONTRIBUTING.md`, `CHANGELOG.md`).

Tests sit next to the code they verify: command tests in `cmd/*_test.go`, package tests in `internal/*/*_test.go`, plus fuzz tests like `internal/git/fuzzer_test.go`.

## Build, Test, and Development Commands
- `make build` builds the local binary to `bin/dcli`.
- `make build-all` cross-compiles release binaries for macOS, Linux, and Windows.
- `make test` runs the full test suite with coverage: `go test -v -cover ./...`.
- `make lint` runs `golangci-lint` with the repo config in `.golangci.yml`.
- `make fuzz` runs the defined fuzz targets for Docker, Git, and config parsing.
- `make install` builds and copies the binary to `~/.local/bin/dcli`.

## Coding Style & Naming Conventions
Use standard Go formatting and keep files `gofmt`-clean. Prefer short, explicit package names and Cobra command files named after the subcommand they implement (`cmd/<area>_<action>.go`). Tests should mirror the target file name. Keep indentation Go-standard with tabs where `gofmt` applies.

Linting is enforced with `gosec`, `errcheck`, `staticcheck`, `gosimple`, `typecheck`, and `ineffassign`. Run `make lint` before opening a PR.

## Testing Guidelines
Write table-driven Go tests where practical. Name tests with the standard Go pattern: `TestXxx` and fuzzers as `FuzzXxx`. Add coverage for both success and failure paths when changing command behavior or config parsing. Run `make test` locally; run `make fuzz` when touching parser or command argument logic.

## Commit & Pull Request Guidelines
Recent history uses concise conventional-style subjects such as `fix: ...`, `ci: ...`, and `chore(deps): ...`. Follow that format and keep the scope meaningful.

PRs should include a clear summary, linked issue when applicable, and terminal output or screenshots for user-facing CLI changes. Note any config, branch-behavior, or platform-specific impact in the description.

## Security & Configuration Tips
User configuration lives in `~/.dcli/config.yaml`. Do not commit machine-specific paths or secrets. If behavior depends on local paths or environment, document it in `README.md` or `docs/CONFIGURATION.md`.
