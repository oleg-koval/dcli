# Security Policy

## Reporting a Vulnerability

**Do not open public GitHub issues for security vulnerabilities.**

If you discover a security vulnerability in dcli, please email security@example.com with:

1. **Description** - What is the vulnerability?
2. **Location** - Which file(s) and line(s)?
3. **Impact** - What could an attacker do?
4. **Reproduction** - Steps to reproduce (if possible)
5. **Fix** - Do you have a suggested fix?

We will:
- Acknowledge receipt within 48 hours
- Investigate and assess severity
- Develop and test a fix
- Release a patch version
- Credit you in the security advisory (unless you prefer anonymity)

## Security Best Practices

### For Users

1. **Keep dcli updated** - Always use the latest version
   ```bash
   brew upgrade dcli
   ```

2. **Validate configuration** - Review `~/.dcli/config.yaml` permissions
   ```bash
   ls -la ~/.dcli/
   chmod 600 ~/.dcli/config.yaml  # Restrict to user only
   ```

3. **Safe repository paths** - Only configure repositories you trust

### For Developers

1. **Input validation** - All paths and commands are validated
2. **Subprocess execution** - Uses `exec.Command` (no shell injection)
3. **Error handling** - Errors are wrapped with context
4. **No network calls** - Local operations only (Docker and Git)
5. **Minimal dependencies** - Only Cobra and yaml.v3

## Security Scanning

dcli uses:
- **OpenSSF Scorecard** - Supply chain security assessment
- **GitHub's Code Scanning** - Static analysis on all commits
- **Automated testing** - 15+ tests across platforms

## Known Limitations

1. **No authentication** - dcli does not implement authentication
2. **File system access** - Requires access to configured repositories and Docker
3. **Docker socket** - Requires access to Docker daemon (usually root)
4. **Git credentials** - Uses system Git configuration (SSH keys, credentials)

## Supported Versions

| Version | Status | Support |
|---------|--------|---------|
| v0.1.x | Current | Actively maintained |

## CVE Disclosure

If a vulnerability is discovered that affects dcli:
1. We will create a security advisory
2. A patch release will be issued
3. The advisory will be published on GitHub
4. Users will be notified

## Responsible Disclosure Timeline

- Day 0: Vulnerability reported
- Day 1: Acknowledgment sent
- Day 7: Fix developed and tested
- Day 10: Patch released and advisory published
- Day 14: Vulnerability disclosure (after patches available)

## Compliance

dcli follows:
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Best Practices](https://go.dev/doc/security)

## Contact

- **Security Issues**: (via private report, no public email here)
- **General Questions**: [GitHub Discussions](https://github.com/oleg-koval/dcli/discussions)
- **Bug Reports**: [GitHub Issues](https://github.com/oleg-koval/dcli/issues)

---

**Thank you for helping keep dcli secure!**
