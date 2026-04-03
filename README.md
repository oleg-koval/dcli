<p align="center">
  <a href="https://github.com/oleg-koval/dcli/actions/workflows/test.yml"><img src="https://github.com/oleg-koval/dcli/actions/workflows/test.yml/badge.svg" alt="tests"></a>
  <a href="https://coveralls.io/github/oleg-koval/dcli"><img src="https://coveralls.io/repos/github/oleg-koval/dcli/badge.svg" alt="Coverage Status"></a>
  <a href="https://goreportcard.com/report/github.com/oleg-koval/dcli"><img src="https://goreportcard.com/badge/github.com/oleg-koval/dcli" alt="Go Report Card"></a>
  <a href="https://securityscorecards.dev/viewer/?uri=github.com/oleg-koval/dcli"><img src="https://api.securityscorecards.dev/projects/github.com/oleg-koval/dcli/badge" alt="OpenSSF Scorecard"></a>
</p>

<h1 align="center">dcli</h1>

<p align="center">
  Lightweight Docker Compose and Git management CLI<br>
  <strong>Clean, restart, and manage repositories with a single command</strong>
</p>

---

## Features

- 🐳 **Docker Management** - Clean containers/volumes, rebuild, and restart services
- 🔄 **Git Batch Operations** - Reset multiple repositories to develop or acceptance branches
- 🚀 **Homebrew Distribution** - Install with a single command: `brew install dcli`
- 🖥️ **Cross-Platform** - Works on macOS (Intel & Apple Silicon), Linux, and Windows
- ⚙️ **YAML Configuration** - Simple config file at `~/.dcli/config.yaml`
- 📝 **Clear Error Messages** - Comprehensive feedback on what went wrong and why
- 🧪 **Well-Tested** - 15+ passing tests across all platforms

## Installation

### Using Homebrew (Recommended)

```bash
brew tap oleg-koval/dcli
brew install dcli
dcli --version
```

### From Source

```bash
git clone https://github.com/oleg-koval/dcli.git
cd dcli
make build
./bin/dcli --version
```

### Direct Download

Download binaries for your platform from [GitHub Releases](https://github.com/oleg-koval/dcli/releases/tag/v0.1.0)

## Quick Start

### Docker Commands

```bash
# Clean all services (remove containers, volumes, rebuild, restart)
dcli docker clean

# Clean specific services
dcli docker clean api web

# Restart services while preserving data
dcli docker restart

# Restart specific services
dcli docker restart api
```

### Git Commands

```bash
# Reset all configured repos to develop
dcli git reset develop

# Reset all configured repos to acceptance
dcli git reset acceptance
```

## Configuration

Create `~/.dcli/config.yaml`:

```yaml
repositories:
  - path: /Users/username/projects/backend
    name: backend
  - path: /Users/username/projects/frontend
    name: frontend
  - path: /Users/username/projects/infra
    name: infra
```

### Environment Variables

- `DCLI_PROJECT_DIR` - Override default project directory (defaults to current directory)

Example:
```bash
DCLI_PROJECT_DIR=/path/to/monorepo dcli docker clean api web
```

## Commands Reference

### Global Flags

- `-h, --help` - Show help
- `-v, --version` - Show version

### Docker Subcommand

```
dcli docker clean [service ...]      # Clean and rebuild (removes containers, volumes, rebuilds)
dcli docker restart [service ...]    # Restart services (preserves data)
```

### Git Subcommand

```
dcli git reset [develop|acceptance]  # Reset all configured repos to specified branch
```

## System Requirements

- **Docker** 20.10+ with Docker Compose
- **Git** 2.20+
- macOS 10.14+, Ubuntu 18.04+, or Windows 10+

## Documentation

- 📖 [Installation Guide](docs/INSTALL.md) - Detailed installation instructions for all platforms
- ⚙️ [Configuration Guide](docs/CONFIGURATION.md) - Complete configuration reference with examples
- 📝 [Contributing](CONTRIBUTING.md) - How to contribute to dcli
- 🔒 [Security Policy](SECURITY.md) - Reporting security vulnerabilities

## Use Cases

### Development Workflow

Reset your working environment to latest develop:
```bash
dcli git reset develop      # Fetch and reset all repos
dcli docker clean           # Clean all containers and volumes
# Fresh environment ready for new feature branch
```

### Quick Service Restart

After code changes or configuration updates:
```bash
dcli docker restart web api  # Restart specific services
# Preserves database data and volumes
```

### Monorepo Management

Configure all microservices and reset with one command:
```bash
# In ~/.dcli/config.yaml: add all repo paths
dcli git reset acceptance   # All services to acceptance branch
dcli docker clean           # Clean all microservices
```

## Architecture

dcli is built with:
- **Go 1.21+** - Compiled language for reliability
- **Cobra** - Battle-tested CLI framework
- **YAML** - Human-readable configuration
- **Docker Compose API** - Direct execution without shells

Zero external dependencies for core functionality.

## Project Status

- ✅ **Alpha Release** (v0.1.0)
- ✅ Tests passing (15+ tests)
- ✅ Cross-platform builds (macOS, Linux, Windows)
- ✅ Homebrew distribution ready
- 🚀 Production-ready for Docker Compose and Git workflows

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details

## Author

[@oleg-koval](https://github.com/oleg-koval)

---

<p align="center">
  <strong>dcli makes container and repository management effortless</strong><br>
  <a href="https://github.com/oleg-koval/dcli/issues">Report Issues</a> • 
  <a href="https://github.com/oleg-koval/dcli/discussions">Discussions</a> •
  <a href="https://github.com/oleg-koval/dcli/releases">Releases</a>
</p>
