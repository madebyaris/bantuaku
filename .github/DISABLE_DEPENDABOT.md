# Disabling Dependabot

Dependabot has been disabled for this repository. We use **CodeRabbit** for code reviews and manage dependencies manually.

## ✅ What Was Done

1. **Configuration File**: Renamed `.github/dependabot.yml` → `.github/dependabot.yml.disabled`
2. **Documentation Updated**: Removed Dependabot references from README and security guides
3. **GitHub Settings**: You need to disable Dependabot in GitHub settings (see below)

## 🔧 Disable Dependabot in GitHub

To completely disable Dependabot:

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Code security and analysis**
3. Under **Dependabot alerts**, click **Disable** (if enabled)
4. Under **Dependabot security updates**, click **Disable** (if enabled)

**Note**: You can keep **Dependabot alerts** enabled if you want to see vulnerability notifications without automatic PRs.

## 📋 Handling Existing Dependabot PRs

You have several options for existing Dependabot pull requests:

### Option 1: Close All (Recommended)
- Go to **Pull Requests** → Filter by `dependencies` label
- Select all Dependabot PRs
- Click **Close** (you can reopen later if needed)

### Option 2: Review and Merge Important Ones
- Review security-related updates first
- Merge critical updates manually
- Close the rest

### Option 3: Bulk Close via GitHub CLI
```bash
# Install GitHub CLI if needed
brew install gh

# Authenticate
gh auth login

# Close all Dependabot PRs
gh pr list --author dependabot --json number --jq '.[].number' | xargs -I {} gh pr close {}
```

## 🔄 Re-enabling Dependabot (If Needed)

If you want to re-enable Dependabot in the future:

1. Rename the config file back:
   ```bash
   mv .github/dependabot.yml.disabled .github/dependabot.yml
   ```

2. Enable in GitHub Settings:
   - **Settings** → **Code security and analysis**
   - Enable **Dependabot alerts** and **Dependabot security updates**

3. Commit and push the changes

## 🛡️ Alternative: Manual Dependency Management

Without Dependabot, manage dependencies manually:

### Frontend (npm)
```bash
cd frontend

# Check for outdated packages
npm outdated

# Check for security vulnerabilities
npm audit

# Update specific package
npm update <package-name>

# Update all packages (careful!)
npm update
```

### Backend (Go)
```bash
cd backend

# Check for outdated modules
go list -u -m all

# Update specific module
go get -u <module-path>

# Update all modules (careful!)
go get -u ./...
```

### Security Scanning
- Use `npm audit` for frontend
- Use `govulncheck` for backend
- CodeRabbit will review dependency changes in PRs
- CodeQL will scan for security issues

## 📝 Why CodeRabbit Instead?

- **Comprehensive Reviews**: Reviews code quality, security, and best practices
- **Context-Aware**: Understands your entire codebase
- **Natural Language**: Ask questions about code changes
- **Less Noise**: No automatic PR spam
- **Better Control**: Review dependencies when you choose to update them

---

**Last Updated**: January 2025
