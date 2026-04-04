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

### DCLI_DISABLE_AUTO_UPDATE

Disable the GitHub Releases self-update check that runs on launch:

```bash
DCLI_DISABLE_AUTO_UPDATE=1 dcli git reset develop
```

### DCLI_AUTO_UPDATE_TIMEOUT

Override the best-effort startup update timeout with a Go duration string:

```bash
DCLI_AUTO_UPDATE_TIMEOUT=250ms dcli git reset develop
```

The default timeout is `1s`.

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
