# Configuration Reference

Complete reference for all available configuration options.

<br/>

## Table of Contents

- [Required Inputs](#required-inputs)
- [Optional Inputs](#optional-inputs)
  - [Commit Settings](#commit-settings)
  - [Tag Settings](#tag-settings)
  - [Pull Request Settings](#pull-request-settings)
- [Default Values](#default-values)

---

## Required Inputs

| Input | Description | Example |
|-------|-------------|---------|
| `user_email` | Git user email | `actions@github.com` |
| `user_name` | Git user name | `GitHub Actions` |

---

## Optional Inputs

<br/>

### Commit Settings

| Input | Description | Default |
|-------|-------------|---------|
| `commit_message` | Commit message | `Auto commit by Go Git Commit Action` |
| `branch` | Branch to push to | `main` |
| `repository_path` | Path to the repository | `.` |
| `file_pattern` | File pattern to add | `.` |
| `skip_if_empty` | Skip if no changes | `false` |

**Notes:**
- `file_pattern` supports multiple space-separated patterns: `"*.md *.txt"`
- `repository_path` is relative to the workspace root

<br/>

### Tag Settings

| Input | Description | Default |
|-------|-------------|---------|
| `tag_name` | Tag name to create or delete | - |
| `tag_message` | Tag message (for annotated tags) | - |
| `delete_tag` | Whether to delete the tag | `false` |
| `tag_reference` | Git reference for the tag | - |

**Notes:**
- Tag operations only execute when `tag_name` is provided
- `tag_reference` can be a commit SHA, tag name, or branch name
- `tag_reference` cannot be used with `delete_tag`

<br/>

### Pull Request Settings

| Input | Description | Default |
|-------|-------------|---------|
| `create_pr` | Whether to create a pull request | `false` |
| `auto_branch` | Whether to create automatic branch | `false` |
| `pr_title` | Pull request title | `Auto PR by Go Git Commit Action` |
| `pr_base` | Base branch for pull request | `main` |
| `pr_branch` | Branch to create pull request from | - |
| `delete_source_branch` | Delete source branch after PR | `false` |
| `github_token` | GitHub token for PR creation | - |
| `pr_labels` | Labels (comma-separated) | - |
| `pr_body` | Custom body message | - |
| `pr_closed` | Close PR after creation | `false` |
| `pr_dry_run` | Simulate PR creation | `false` |

**Notes:**
- `github_token` is required when `create_pr` is true
- `pr_branch` is required when `create_pr` is true and `auto_branch` is false
- `pr_base` is required when `create_pr` is true
- When `auto_branch` is true, creates branch with format: `update-files-{timestamp}`
- `delete_source_branch` only works with `auto_branch: true`

---

## Default Values

```yaml
commit_message: "Auto commit by Go Git Commit Action"
branch: "main"
repository_path: "."
file_pattern: "."
skip_if_empty: false
delete_tag: false
create_pr: false
auto_branch: false
pr_base: "main"
delete_source_branch: false
pr_closed: false
pr_dry_run: false
```

---

## Input Validation

The action validates inputs before execution:

<br/>

### PR Creation Validation
- `pr_branch` must be set when `auto_branch` is false
- `pr_base` must be set when `create_pr` is true
- `github_token` must be set when `create_pr` is true

### Tag Validation
- `tag_reference` cannot be used with `delete_tag`

### File Pattern Validation
- Multiple patterns separated by spaces
- Standard git pattern matching supported
