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

## Custom Command Packs

Custom commands are stored in JSON pack files:

- Local personal pack: `~/.dcli/commands.json`
- Repo-shared pack: `./.dcli/commands.json` at the repository root

Shared packs load automatically when you run dcli inside a repository that contains a pack file. Local packs are always loaded from your home directory.

Use `dcli commands` to inspect loaded commands, `dcli commands add` to create a personal shortcut, and `dcli commands export` / `dcli commands import` to move packs between developers.
Use `dcli <command>` for normal execution. The command-management surface lives under `dcli commands`, and the interactive browser is `dcli commands ui`.

For an interactive browser, use:

```bash
dcli commands ui --export-file ./team-pack.json
```

Inside the UI:

- `j/k` or arrow keys move the cursor
- `space` selects commands for export
- `enter` or `r` runs the highlighted command
- `e` toggles enable/disable
- `d` deletes the highlighted command
- `x` exports the selected commands to the file passed with `--export-file`
- `q` quits

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

### DCLI_AUTO_UPDATE_CHANNEL

Choose which GitHub Releases line the startup updater may install (default: stable only).

| Value | Behavior |
|-------|----------|
| `stable` or unset | Latest **non-prerelease** GitHub release (same as before). |
| `prerelease` or `pre` | Newest semver among **prerelease** GitHub releases with a matching binary. |
| `beta` | Like `prerelease`, but only tags containing `-beta` (case-insensitive). |
| `alpha` | Like `prerelease`, but only tags containing `-alpha`. |

Unknown values fall back to `stable`.

```bash
DCLI_AUTO_UPDATE_CHANNEL=prerelease dcli docker clean
```

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
