# Security Scanning Setup Guide

This document explains how to enable and configure security scanning tools for the Bantuaku repository.

## 🔧 GitHub Settings Configuration

**Note**: Dependabot is disabled. We use CodeRabbit for code reviews and manage dependencies manually.

### Step 1: Enable Code Scanning (CodeQL)

1. Go to **Settings** → **Code security and analysis**
2. Under **Code scanning**, find **CodeQL analysis**
3. Click **Set up** or **Enable**
4. Select **Default** setup (or use the existing workflow file)

The workflow file `.github/workflows/codeql.yml` is already configured and will:
- Run on every push to main/master/develop branches
- Run on pull requests
- Run weekly on Mondays
- Analyze both JavaScript (frontend) and Go (backend) code

### Step 2: Enable Secret Scanning

1. Go to **Settings** → **Code security and analysis**
2. Under **Secret scanning**, click **Enable**

This will scan your repository for accidentally committed secrets (API keys, passwords, etc.)

### Step 3: Enable Dependency Review

1. Go to **Settings** → **Code security and analysis**
2. Under **Dependency review**, click **Enable**

This will automatically review dependencies in pull requests.

## 📋 Workflow Files

All security workflows are located in `.github/workflows/`:

### `codeql.yml`
- **Purpose**: Static code analysis for security vulnerabilities
- **Triggers**: Push, PR, Weekly schedule
- **Languages**: JavaScript, Go
- **Config**: Uses `.github/codeql-config.yml`

### `dependency-review.yml`
- **Purpose**: Review dependencies in pull requests
- **Triggers**: Pull requests to main/master/develop
- **Action**: Blocks PRs with moderate+ severity vulnerabilities

### `security.yml`
- **Purpose**: Additional security scans (npm audit, Gosec, govulncheck)
- **Triggers**: Push, PR, Weekly schedule
- **Scans**: 
  - Frontend: npm audit
  - Backend: Gosec, govulncheck

## 📊 Viewing Results

### Code Scanning Results
- Go to **Security** tab → **Code scanning**
- View all CodeQL findings
- Filter by severity, language, etc.

### Dependency Review
- View directly in pull requests
- Appears as a check on PRs

## 🔔 Notifications

Configure notifications:
1. Go to **Settings** → **Notifications**
2. Enable notifications for:
   - Security alerts
   - Code scanning alerts
   - CodeRabbit reviews

## 🛠️ Local Testing

### Test Workflow Files
```bash
# Install act (local GitHub Actions runner)
brew install act  # macOS
# or download from https://github.com/nektos/act

# Test CodeQL workflow (dry run)
act -W .github/workflows/codeql.yml --dry-run
```

### Test Security Scans Locally

**Frontend:**
```bash
cd frontend
npm audit
npm audit --audit-level=moderate
```

**Backend:**
```bash
cd backend
# Install Gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run Gosec
gosec ./...

# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run govulncheck
govulncheck ./...
```

## 📝 Best Practices

1. **Review CodeRabbit Suggestions Regularly**
   - Check CodeRabbit reviews on pull requests
   - Address security and quality suggestions
   - Use CodeRabbit chat for clarifications

2. **Manage Dependencies Manually**
   - Review dependency updates when needed
   - Test updates before merging
   - Use `npm audit` and `govulncheck` for security checks

3. **Address Security Alerts Promptly**
   - Critical/High: Fix within 7 days
   - Moderate: Fix within 30 days
   - Low: Fix when convenient

3. **Monitor Code Scanning Results**
   - Review new findings weekly
   - Fix false positives by adding suppressions
   - Track security debt

4. **Keep Workflows Updated**
   - Update GitHub Actions versions monthly
   - Review and update CodeQL queries
   - Keep security tools up to date

## 🚨 Troubleshooting

### CodeQL Not Running
- Verify workflow file syntax
- Check repository Actions settings
- Ensure CodeQL is enabled in repository settings

### False Positives
- Add suppressions to CodeQL config
- Use inline comments to suppress specific findings
- Document why suppression is needed

## 📚 Resources

- [GitHub Security Documentation](https://docs.github.com/en/code-security)
- [CodeQL Documentation](https://docs.github.com/en/code-security/code-scanning)
- [CodeRabbit Documentation](https://docs.coderabbit.ai)
- [Security Policy Template](https://github.com/github/security-policy-template)

---

**Last Updated**: January 2025
