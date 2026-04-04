# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**dcli** is a lightweight Docker Compose and Git management CLI written in Go. It allows users to clean/restart Docker services and batch reset Git repositories across a multi-repository setup.

**Stack:** Go 1.25.0+, Cobra CLI framework, YAML configuration
**Status:** Alpha release (v0.1.0)
**Key Dependencies:** Cobra, gopkg.in/yaml.v3, and `github.com/creativeprojects/go-selfupdate v1.5.2` for startup updates
**Approved Exception:** `github.com/creativeprojects/go-selfupdate v1.5.2` is intentionally included for the auto-update feature. Maintainers approved this exception to the "Cobra and gopkg.in/yaml.v3" dependency guideline.

## Quick Commands

| Task | Command |
|------|---------|
| Build | `make build` |
| Run tests | `make test` |
| Run linter | `make lint` |
| Build all platforms | `make build-all` |
| Install locally | `make install` |
| Clean build artifacts | `make clean` |

### Test-Related Commands

- **Run all tests:** `make test`
- **Run tests for specific package:** `go test -v ./internal/config ./cmd`
- **Run single test:** `go test -v -run TestName ./cmd`
- **Coverage report:** `go test -v -cover ./...`
- **Coverage with output file:** `go test -v -covermode=atomic -coverprofile=coverage.out ./...`
- **Run fuzzer tests:** `make fuzz` (runs fuzzing for 10s on each test)
- **Run linter (SAST):** `make lint` (static analysis with golangci-lint)
- **Verify dependencies:** `go mod verify` (checks go.mod and go.sum are in sync)

## Architecture

```
dcli/
├── main.go                          # Entry point
├── cmd/                             # CLI command definitions
│   ├── root.go                      # Root command with version
│   ├── docker.go                    # Docker command group
│   ├── docker_clean.go              # Docker clean subcommand
│   ├── docker_restart.go            # Docker restart subcommand
│   ├── git.go                       # Git command group
│   └── git_reset.go                 # Git reset subcommand
├── internal/
│   ├── config/                      # Configuration loading
│   │   ├── config.go                # Config file parsing
│   │   └── config_test.go           # Config tests
│   ├── docker/                      # Docker operations
│   │   ├── helpers.go               # Docker Compose API wrappers
│   │   └── helpers_test.go          # Docker tests
│   └── git/                         # Git operations
│       ├── helpers.go               # Git command execution
│       └── helpers_test.go          # Git tests
```

### Key Design Decisions

1. **Cobra for CLI:** Standard Go CLI framework provides command structure and automatic help/versioning
2. **Minimal Dependencies:** Core CLI code stays on Cobra and YAML; startup updates use `github.com/creativeprojects/go-selfupdate v1.5.2`
3. **Config at ~/.dcli/config.yaml:** User's home directory for easy discovery and cross-platform support
4. **Package Organization:** 
   - `cmd/` contains all CLI commands (test in same package with _test.go suffix)
   - `internal/` contains domain logic separated by concern (config, docker, git)
   - Each domain has a helpers file with core logic + corresponding test file

## Common Development Tasks

### Adding a New Command

1. Create command file in `cmd/` (e.g., `cmd/my_command.go`)
2. Define command with Cobra: `var myCmd = &cobra.Command{...}`
3. Add init function to register it: `func init() { rootCmd.AddCommand(myCmd) }`
4. Add tests in `cmd/my_command_test.go`

### Adding Domain Logic

1. Create helper in `internal/{domain}/helpers.go`
2. Test in `internal/{domain}/helpers_test.go`
3. Import and call from `cmd/` command file

### Testing Notes

- Tests use standard Go testing package
- Test coverage is tracked (currently 76%+, with core packages at 90%+)
- Use `-run` flag to run specific tests: `go test -v -run TestDockerClean ./...`
- CI runs against Go 1.25 and stable

## Configuration

User configuration file location: `~/.dcli/config.yaml`

```yaml
repositories:
  - path: /path/to/repo
    name: repo-name
    remote: origin  # optional
```

The config is loaded via `internal/config/config.go` which:
- Returns empty config if file doesn't exist
- Returns error if YAML is malformed
- Stores repositories as list of path/name pairs

## Build System

**Makefile targets:**
- `make build` - Build single binary for current OS/arch
- `make build-all` - Cross-compile for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- Version is set via `-ldflags="-X 'github.com/oleg-koval/dcli/cmd.Version=...'"` at build time

## Platform Support

- macOS 10.14+ (Intel & Apple Silicon)
- Linux (Ubuntu 18.04+)
- Windows 10+

## CI/CD

GitHub Actions workflow (`.github/workflows/test.yml`):
- Runs on push to main/develop and PR to main/develop
- Tests against Go 1.21 and 1.22
- Uploads coverage to Coveralls (atomicmode)

## Known Gotchas

1. **No verbose output by default:** Docker and Git operations may appear silent; look for error messages
2. **DCLI_PROJECT_DIR environment variable:** Overrides current directory for Docker Compose project detection
3. **Docker Compose version:** Requires Docker with Compose support (20.10+)
4. **Git branch names:** Reset command accepts only "develop" or "acceptance"

## Testing Coverage

59 tests across:
- `cmd/` - Command behavior and integration
- `internal/config/` - Configuration parsing
- `internal/docker/` - Docker Compose operations
- `internal/git/` - Git repository operations

Goal: Maintain 85%+ overall coverage (core packages at 90%+)

## Security & Quality Infrastructure

### Static Analysis (SAST)
- **Tool:** golangci-lint with `.golangci.yml` configuration
- **Checks:** gosec (security), errcheck, staticcheck, ineffassign, exhaustive, and more
- **Run:** `make lint` or included in CI/CD pipeline
- **CI Integration:** Automatically runs on all PRs and commits to main/develop

### Fuzzing Integration
- **Location:** `internal/{docker,git,config}/fuzzer_test.go`
- **Purpose:** Tests for edge cases and potential crashes with arbitrary input
- **Run:** `make fuzz` (default 10s per test)
- **Target Areas:**
  - Docker command parsing
  - Git branch name validation
  - YAML config parsing

### Dependency Management
- **Pinned Dependencies:** All dependencies tracked in `go.mod` with explicit versions
- **Build Tools:** Tracked in `tools.go` to ensure reproducible builds
- **Automated Updates:** Dependabot configuration at `.github/dependabot.yml`
  - Weekly updates for Go modules and GitHub Actions
  - Automatic PRs with dependency updates
  - Includes transitive dependencies

### Supply Chain Security
- **OpenSSF Scorecard:** Runs weekly to assess security posture
- **GitHub CodeQL:** Integrated for vulnerability scanning (via a dedicated GitHub CodeQL workflow in CI)
- **Dependency Verification:** `go mod verify` ensures integrity in CI

## For Future Development

- Versioning is in `cmd/root.go` (var Version)
- Consider config validation enhancements
- Docker output handling could be improved for better error messages
- Git operations could support additional branch strategies
- Monitor OpenSSF Scorecard results to address recommendations
- Keep fuzzer tests updated as new edge cases are discovered
