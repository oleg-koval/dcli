# Contributing to dcli

Thank you for considering contributing to dcli! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person

## Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Git
- Make

### Setup

```bash
git clone https://github.com/oleg-koval/dcli.git
cd dcli
make build
./bin/dcli --help
```

### Running Tests

```bash
make test
```

## Making Changes

### Branch Naming

- `feature/` - New features (e.g., `feature/add-docker-prune`)
- `fix/` - Bug fixes (e.g., `fix/config-parsing-issue`)
- `docs/` - Documentation updates
- `chore/` - Maintenance tasks

### Commit Messages

Use conventional commits:

```
type(scope): subject

description...

- optional: breaking change info
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Test additions/changes
- `chore` - Build, CI, dependencies
- `refactor` - Code reorganization

Example:
```
feat(docker): add image cleanup option

Allows users to remove Docker images in addition to containers
and volumes when running dcli docker clean.

Breaking change: none
```

### Code Style

- Follow Go conventions (`gofmt`, `go vet`)
- Use descriptive variable names
- Add comments for unexported functions
- Keep functions focused and small

### Testing

- Write tests for new functionality
- Maintain existing test coverage (currently 15+ tests)
- Test across platforms (macOS, Linux, Windows)

```bash
go test -v ./...
```

## Pull Request Process

1. Fork the repository
2. Create feature branch: `git checkout -b feature/description`
3. Make your changes
4. Add/update tests
5. Run tests locally: `make test`
6. Commit with conventional message
7. Push to fork
8. Create Pull Request with description of changes
9. Address review comments

### PR Title Format

```
type(scope): brief description
```

Same format as commits above.

### PR Description Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Documentation update
- [ ] Refactoring

## Testing
Describe how you tested the changes.

## Related Issues
Closes #123

## Checklist
- [ ] Tests pass locally
- [ ] No breaking changes
- [ ] Documentation updated
```

## Reporting Issues

### Bug Reports

Include:
- OS and version
- Go version
- Docker version
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

### Feature Requests

Include:
- Use case/motivation
- Proposed solution
- Alternatives considered
- Any additional context

## Project Roadmap

Current priorities:
1. Core Docker and Git functionality (✅ done)
2. Cross-platform support (✅ done)
3. Homebrew distribution (✅ done)
4. Enhanced error messages
5. Shell completion scripts
6. Configuration templating

## License

By contributing, you agree your work is licensed under MIT.

## Questions?

- Open a [GitHub Discussion](https://github.com/oleg-koval/dcli/discussions)
- Check existing [Issues](https://github.com/oleg-koval/dcli/issues)

---

**Thank you for contributing!**
