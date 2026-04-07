# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.4] - 2026-04-08

### Added
- **Docker Compose profile support**: Use `--profile` flag to target services
  in specific Docker Compose profiles (e.g., `dcli docker --profile all_services restart`).
  Supports multiple profiles.

### Fixed
- Docker commands now discover all services, not just those in the default profile
- Docker build and restart output is now visible in the terminal (previously silent)
- Project directory resolution uses the actual working directory instead of a
  relative `"."` path, fixing issues when dcli is invoked from scripts or aliases

### Improved
- `BuildCleanCommandArgs` and `BuildRestartCommandArgs` helpers are now used
  consistently, eliminating duplicated argument construction

## [0.1.0] - 2026-04-03

### Added
- Initial release
- Docker commands: `clean`, `restart`
- Git commands: `reset` (develop, acceptance)
- Config file support (`~/.dcli/config.yaml`)
- GitHub Actions CI/CD
- Homebrew distribution
- Cross-platform builds (macOS, Linux, Windows)
