# Authentication Guide

This guide covers authentication requirements and best practices for using the Go Git Commit Action.

<br/>

## Table of Contents

- [GitHub Token for checkout@v6](#github-token-for-checkoutv6)
- [Token Types](#token-types)
- [Authentication Scenarios](#authentication-scenarios)
- [Best Practices](#best-practices)

---

## GitHub Token for checkout@v6

<br/>

### Important Notice

Starting from `actions/checkout@v6`, Git credentials are stored in the `$RUNNER_TEMP` directory, which is **not accessible from Docker containers**. You must provide the `github_token` input:

```yaml
- uses: actions/checkout@v6

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    github_token: ${{ secrets.PAT_TOKEN }}  # Required!
```

<br/>

### Why is this needed?

1. `checkout@v6` changed credential storage for security improvements
2. Docker containers cannot access `$RUNNER_TEMP` directory
3. The action configures Git credentials using the provided token

---

## Token Types

<br/>

### GITHUB_TOKEN (Default)

Automatically provided by GitHub Actions. Use for:
- ✅ Standard commit/push operations
- ✅ Basic branch operations
- ❌ Tag operations (limited permissions)
- ❌ Workflow modifications (limited permissions)

```yaml
github_token: ${{ secrets.GITHUB_TOKEN }}
```

<br/>

### Personal Access Token (PAT)

User-created token with custom permissions. Use for:
- ✅ Tag creation and deletion
- ✅ Workflow file modifications
- ✅ Cross-repository operations
- ✅ Pull request creation

```yaml
github_token: ${{ secrets.PAT_TOKEN }}
```

**Required PAT Permissions:**
- `repo` (Full control of private repositories)
- `workflow` (Update GitHub Action workflows) - if modifying workflows

---

## Authentication Scenarios

<br/>

### Scenario 1: Simple Commit and Push

```yaml
- uses: actions/checkout@v6
  with:
    token: ${{ secrets.GITHUB_TOKEN }}

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    github_token: ${{ secrets.GITHUB_TOKEN }}  # Required for checkout@v6
```

<br/>

### Scenario 2: Tag Operations

```yaml
- uses: actions/checkout@v6
  with:
    token: ${{ secrets.PAT_TOKEN }}  # PAT required for tags

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    tag_name: v1.0.0
    github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### Scenario 3: Pull Request Creation

```yaml
- uses: actions/checkout@v6
  with:
    token: ${{ secrets.GITHUB_TOKEN }}

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature
    pr_base: main
    github_token: ${{ secrets.GITHUB_TOKEN }}  # Required for PR
```

<br/>

### Scenario 4: Using checkout@v6

```yaml
- uses: actions/checkout@v6
  with:
    token: ${{ secrets.PAT_TOKEN }}

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    github_token: ${{ secrets.PAT_TOKEN }}  # Always required with v6
```

---

## Best Practices

<br/>

### 1. Token Selection

| Operation | Recommended Token |
|-----------|-------------------|
| Commit & Push | `GITHUB_TOKEN` |
| Tag Management | `PAT_TOKEN` |
| PR Creation | `GITHUB_TOKEN` or `PAT_TOKEN` |
| Workflow Modifications | `PAT_TOKEN` |

<br/>

### 2. Token Security

✅ **Do:**
- Store tokens as repository secrets
- Use the least privileged token for the job
- Rotate PATs regularly

❌ **Don't:**
- Hardcode tokens in workflow files
- Use PAT when GITHUB_TOKEN is sufficient
- Share tokens across different repositories unnecessarily

<br/>

### 3. Checkout Version

| Version | Token Required | Notes |
|---------|----------------|-------|
| checkout@v5 | Optional | Works without explicit token for basic operations |
| checkout@v6 | **Required** | Must provide token for Docker actions |

<br/>

### 4. Error Handling

If you encounter authentication errors:

1. **Verify token is provided** when using checkout@v6
2. **Check token permissions** for tag/workflow operations
3. **Ensure token is not expired** for PATs
4. **Verify secret name** matches in workflow file

---

## Creating a Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Click "Generate new token (classic)"
3. Select scopes:
   - ✅ `repo` - Full control of private repositories
   - ✅ `workflow` - Update GitHub Action workflows (if needed)
4. Generate and copy the token
5. Add to repository secrets as `PAT_TOKEN`

---

## Troubleshooting

<br/>

### Error: "Authentication failed"

**Solution:** Provide `github_token` input, especially with checkout@v6

```yaml
github_token: ${{ secrets.GITHUB_TOKEN }}
```

<br/>

### Error: "Insufficient permissions for tag operations"

**Solution:** Use PAT instead of GITHUB_TOKEN

```yaml
# In checkout step
token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### Error: "Could not read from remote repository"

**Solution:** Ensure token has `repo` scope
