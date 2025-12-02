# CodeRabbit Setup Guide

This guide will help you set up CodeRabbit AI code reviews for the Bantuaku repository.

## 🚀 Quick Start

### Step 1: Install CodeRabbit GitHub App

1. Go to [CodeRabbit.ai](https://coderabbit.ai)
2. Click **"Get Started"** or **"Install GitHub App"**
3. Sign in with your GitHub account
4. Select the **Bantuaku** repository (`madebyaris/bantuaku`)
5. Choose installation options:
   - **All repositories** (if you want it everywhere)
   - **Only select repositories** → Choose `bantuaku`
6. Click **"Install"** or **"Save"**

### Step 2: Configure Repository Access

CodeRabbit needs access to:
- ✅ **Pull requests** - To review code
- ✅ **Contents** - To read repository files
- ✅ **Metadata** - To understand repository structure
- ✅ **Pull request reviews** - To post review comments

These permissions are automatically requested during installation.

### Step 3: Verify Configuration File

The repository already includes `.coderabbit.yaml` with optimized settings for:
- **Go backend** (`backend/**/*.go`)
- **React/TypeScript frontend** (`frontend/src/**/*.{ts,tsx}`)
- **Security scanning** (Gitleaks, ESLint)
- **Path filters** (excludes build artifacts, dependencies)

You can customize this file if needed (see [Configuration Options](#configuration-options) below).

### Step 4: Test CodeRabbit

1. Create a test pull request:
   ```bash
   git checkout -b test-coderabbit
   # Make a small change
   git commit -m "test: verify CodeRabbit integration"
   git push origin test-coderabbit
   ```
2. Open a pull request on GitHub
3. CodeRabbit will automatically:
   - Review your code
   - Post a summary comment
   - Provide line-by-line suggestions
   - Answer questions in PR comments

## 📋 Configuration Details

### Current Configuration (`.coderabbit.yaml`)

#### Review Profile
- **Profile**: `chill` - Standard review depth (change to `assertive` for more detailed feedback)
- **Auto-review**: Enabled for all PRs
- **Request changes**: Disabled (only suggestions, not blocking)

#### Path Filters
CodeRabbit focuses on:
- ✅ `backend/**/*.go` - Go source files
- ✅ `frontend/src/**/*.{ts,tsx,js,jsx}` - Frontend source files
- ❌ Excludes: `node_modules`, `dist`, `build`, `vendor`, test files, etc.

#### Language-Specific Instructions

**Go Backend:**
- Error handling best practices
- Security (SQL injection, authentication)
- Concurrency (goroutines, channels)
- Resource management (defer statements)
- API design (RESTful conventions)

**React/TypeScript Frontend:**
- React hooks and component structure
- TypeScript type safety
- Security (XSS, CSRF)
- Performance (memoization, lazy loading)
- Accessibility (a11y)
- State management (Zustand)

#### Integrated Tools
- **ESLint** - JavaScript/TypeScript linting
- **Gitleaks** - Secret detection
- **Golangci** - Go linting (if available)

## 🔧 Configuration Options

### Change Review Profile

Edit `.coderabbit.yaml`:

```yaml
reviews:
  profile: "assertive"  # More detailed reviews
  # or
  profile: "chill"      # Standard reviews (current)
```

### Enable Auto-Request Changes

```yaml
reviews:
  request_changes_workflow: true  # CodeRabbit can request changes
```

### Custom Path Instructions

Add specific instructions for certain files:

```yaml
reviews:
  path_instructions:
    - path: "backend/handlers/**/*.go"
      instructions: |
        Focus on API security, input validation, and error handling.
        Ensure all endpoints have proper authentication middleware.
```

### Disable Auto-Review

```yaml
reviews:
  auto_review:
    enabled: false  # Manual review trigger only
```

## 💬 Using CodeRabbit Chat

CodeRabbit can answer questions in PR comments:

1. **Ask questions** in PR comments:
   ```
   @coderabbitai Can you explain this function?
   @coderabbitai Is this secure?
   @coderabbitai How can I optimize this?
   ```

2. **Request specific reviews**:
   ```
   @coderabbitai Please review the security implications
   @coderabbitai Check for performance issues
   ```

3. **Get explanations**:
   ```
   @coderabbitai Why did you suggest this change?
   ```

## 🎯 Best Practices

### For Pull Requests

1. **Keep PRs focused** - Smaller PRs get better reviews
2. **Add context** - Use PR descriptions to help CodeRabbit understand changes
3. **Respond to suggestions** - Engage with CodeRabbit's feedback
4. **Ask questions** - Use chat feature for clarifications

### For Code Quality

1. **Review suggestions** - CodeRabbit catches many issues early
2. **Learn from feedback** - Use suggestions to improve coding skills
3. **Customize rules** - Adjust `.coderabbit.yaml` for your team's preferences
4. **Combine with other tools** - CodeRabbit complements CodeQL and Dependabot

## 🔍 Troubleshooting

### CodeRabbit Not Reviewing PRs

1. **Check installation**: Go to repository Settings → Integrations → CodeRabbit
2. **Verify permissions**: Ensure CodeRabbit has access to pull requests
3. **Check configuration**: Verify `.coderabbit.yaml` syntax is valid
4. **Review logs**: Check CodeRabbit's activity in the PR

### Too Many Suggestions

1. **Adjust profile**: Change to `chill` in `.coderabbit.yaml`
2. **Narrow path filters**: Exclude more paths if needed
3. **Customize instructions**: Add specific guidelines to reduce noise

### Missing Reviews

1. **Check path filters**: Ensure your files match the filters
2. **Verify file types**: CodeRabbit focuses on configured file types
3. **Enable auto-review**: Ensure `auto_review.enabled: true`

## 📚 Additional Resources

- [CodeRabbit Documentation](https://docs.coderabbit.ai)
- [Configuration Reference](https://docs.coderabbit.ai/reference/configuration)
- [Cursor IDE Extension](https://docs.coderabbit.ai/integrations/cursor) - For in-editor reviews
- [CodeRabbit GitHub](https://github.com/coderabbitai/ai-pr-reviewer)

## 🎨 Cursor IDE Integration (Optional)

For real-time code reviews in Cursor:

1. Install CodeRabbit extension in Cursor
2. Sign in with your GitHub account
3. CodeRabbit will review code as you type
4. Get instant feedback without creating PRs

**Note**: This requires a CodeRabbit account (free tier available).

## ✅ Verification Checklist

After setup, verify:

- [ ] CodeRabbit GitHub App is installed
- [ ] Repository is selected in CodeRabbit settings
- [ ] `.coderabbit.yaml` exists in repository root
- [ ] Test PR receives CodeRabbit review
- [ ] Badge appears in README (optional)
- [ ] Team members can see CodeRabbit comments

## 🆘 Support

- **CodeRabbit Issues**: [GitHub Issues](https://github.com/coderabbitai/ai-pr-reviewer/issues)
- **Documentation**: [docs.coderabbit.ai](https://docs.coderabbit.ai)
- **Community**: [Discord](https://discord.gg/coderabbit) (if available)

---

**Last Updated**: January 2025
