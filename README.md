# dcli

A lightweight CLI tool for managing Docker Compose services and Git repositories.

## Features

- **Docker Management**: Clean, rebuild, and restart services
- **Git Management**: Reset multiple repositories to develop or acceptance branches
- **Homebrew Distribution**: Simple `brew install` setup
- **Cross-Platform**: Runs on macOS, Linux, and Windows

## Quick Start

### Install with Homebrew

```bash
brew tap oleg-koval/dcli
brew install dcli
dcli --version
```

### Install from Source

```bash
git clone https://github.com/oleg-koval/dcli.git
cd dcli
make install
```

## Commands

### Docker

```bash
# Clean all services (remove containers, volumes, images, rebuild, restart)
dcli docker clean

# Clean specific services
dcli docker clean api web

# Restart all services (preserves data)
dcli docker restart

# Restart specific services
dcli docker restart api
```

### Git

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
  - path: /path/to/monorepo
    name: monorepo
  - path: /path/to/backend
    name: backend
  - path: /path/to/frontend
    name: frontend
```

## Environment Variables

- `DCLI_PROJECT_DIR`: Override default project directory (defaults to current directory)

## Requirements

- Docker + Docker Compose
- Git
- Go 1.21+ (for building from source)

## Development

```bash
# Run tests
make test

# Build locally
make build

# Build for all platforms
make build-all
```

## License

MIT

## Author

[@oleg-koval](https://github.com/oleg-koval)
