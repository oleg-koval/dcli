# Configuration Guide

## Config File Location

`~/.dcli/config.yaml`

## Repository Configuration

List all repositories you want to manage with `dcli git reset`:

```yaml
repositories:
  - path: /Users/username/projects/monorepo
    name: monorepo
  - path: /Users/username/projects/backend
    name: backend
  - path: /Users/username/projects/frontend
    name: frontend
```

### Fields

- `path`: Absolute path to the Git repository
- `name`: Human-readable name (used in output)

### Creating Config File

If the file doesn't exist, create it:

```bash
mkdir -p ~/.dcli
cat > ~/.dcli/config.yaml << 'EOF'
repositories:
  - path: /path/to/repo1
    name: repo1
  - path: /path/to/repo2
    name: repo2
EOF
```

## Environment Variables

### DCLI_PROJECT_DIR

Override the default project directory for Docker commands:

```bash
DCLI_PROJECT_DIR=/path/to/project dcli docker clean
```

Without this, dcli uses the current directory.

## Examples

### Docker Compose in subdirectory

```bash
cd /path/to/monorepo
dcli docker clean backend frontend
```

Or:

```bash
DCLI_PROJECT_DIR=/path/to/monorepo dcli docker clean backend frontend
```

### Git with custom repos

Edit `~/.dcli/config.yaml` with your actual paths, then:

```bash
dcli git reset develop
```
