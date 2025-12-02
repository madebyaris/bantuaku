# Security Policy

## Supported Versions

We actively support security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Bantuaku seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **Email**: Send an email to [security@bantuaku.id](mailto:security@bantuaku.id) (if available) or contact the maintainers directly
2. **GitHub Security Advisory**: Use GitHub's private vulnerability reporting feature (if enabled)

### What to Include

Please include the following information in your report:

- **Type of issue** (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- **Full paths of source file(s) related to the manifestation of the issue**
- **The location of the affected source code** (tag/branch/commit or direct URL)
- **Step-by-step instructions to reproduce the issue**
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the issue**, including how an attacker might exploit the issue

This information will help us triage your report more quickly.

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 7 days
- **Updates**: We will keep you informed of our progress every 7-14 days
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days

### Disclosure Policy

- We will credit you for the discovery (unless you prefer to remain anonymous)
- We will not disclose your identity without your permission
- We will work with you to understand and resolve the issue quickly
- We will notify you when the vulnerability is fixed and can be disclosed

## Security Best Practices

### For Users

- Keep your dependencies up to date
- Use strong, unique passwords
- Enable two-factor authentication (when available)
- Regularly review your API keys and tokens
- Follow the principle of least privilege

### For Developers

- Never commit secrets or API keys to the repository
- Use environment variables for sensitive configuration
- Keep dependencies updated (we use Dependabot)
- Review security alerts from GitHub
- Follow secure coding practices
- Run security scans before committing code

## Security Tools

We use the following tools to maintain security:

- **Dependabot**: Automated dependency updates and vulnerability scanning
- **CodeQL**: Static analysis for security vulnerabilities
- **npm audit**: Frontend dependency vulnerability scanning
- **Gosec**: Go security checker
- **govulncheck**: Go vulnerability database checker

## Security Updates

Security updates are released as soon as possible after a vulnerability is discovered and patched. We recommend:

- Subscribing to security advisories
- Keeping your deployment updated
- Monitoring the repository for security releases

## Acknowledgments

We would like to thank the security researchers and community members who help keep Bantuaku secure by responsibly disclosing vulnerabilities.

---

**Last Updated**: January 2025
