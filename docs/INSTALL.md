# Installation Guide

## Homebrew (Recommended)

```bash
brew tap oleg-koval/dcli
brew install dcli
dcli --version
```

To update:
```bash
brew upgrade dcli
```

`dcli` also checks GitHub Releases on launch and updates itself when a newer version is available.
Set `DCLI_DISABLE_AUTO_UPDATE=1` to skip that check.
Set `DCLI_AUTO_UPDATE_TIMEOUT` to tune how long that best-effort check can wait before giving up.

## From Source

### Prerequisites
- Go 1.25.0+
- Git

### Steps

```bash
git clone https://github.com/oleg-koval/dcli.git
cd dcli
make build
make install
```

The binary will be installed to `~/.local/bin/dcli`.

Ensure `~/.local/bin` is in your PATH:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Verify Installation

```bash
dcli --version
dcli --help
```

## Troubleshooting

### Command not found after install

Ensure `~/.local/bin` is in your PATH:
```bash
echo $PATH | grep ".local/bin"
```

If not present, add to your shell profile (~/.zshrc, ~/.bashrc):
```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Docker command fails

Ensure Docker and Docker Compose are installed:
```bash
docker --version
docker compose version
```
